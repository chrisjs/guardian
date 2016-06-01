package dadoo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/guardian/rundmc/runrunc"
	"github.com/cloudfoundry/gunk/command_runner"
	"github.com/kr/logfmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pivotal-golang/lager"
)

type ExecRunner struct {
	dadooPath      string
	runcPath       string
	pidGen         runrunc.UidGenerator
	iodaemonRunner runrunc.ExecRunner
	commandRunner  command_runner.CommandRunner
}

func NewExecRunner(dadooPath, runcPath string, pidGen runrunc.UidGenerator, iodaemonRunner runrunc.ExecRunner, commandRunner command_runner.CommandRunner) *ExecRunner {
	return &ExecRunner{
		dadooPath:      dadooPath,
		runcPath:       runcPath,
		pidGen:         pidGen,
		iodaemonRunner: iodaemonRunner,
		commandRunner:  commandRunner,
	}
}

func (d *ExecRunner) Run(log lager.Logger, spec *specs.Process, processesPath, handle string, tty *garden.TTYSpec, pio garden.ProcessIO) (p garden.Process, theErr error) {
	if !contains(spec.Env, "USE_DADOO=true") {
		return d.iodaemonRunner.Run(log, spec, processesPath, handle, tty, pio)
	}

	log = log.Session("exec")

	pid := d.pidGen.Generate()
	path := filepath.Join(processesPath, pid)

	encodedSpec, err := json.Marshal(spec)
	if err != nil {
		return nil, err // this could *almost* be a panic: a valid spec should always encode (but out of caution we'll error)
	}

	if err := os.MkdirAll(path, 0700); err != nil {
		return nil, err
	}

	pipes, pipeArgs, err := mkFifos(pio, filepath.Join(path, "stdin"), filepath.Join(path, "stdout"), filepath.Join(path, "stderr"))
	if err != nil {
		return nil, err
	}

	fd3r, fd3w, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	defer fd3r.Close()

	logFile := filepath.Join(path, "log")
	cmd := exec.Command(d.dadooPath, append(pipeArgs, "-log", logFile, "exec", d.runcPath, path, handle)...)
	cmd.Stdin = bytes.NewReader(encodedSpec)
	cmd.ExtraFiles = []*os.File{
		fd3w,
	}

	if err := d.commandRunner.Start(cmd); err != nil {
		return nil, err
	}

	fd3w.Close()

	if err := pipes.start(); err != nil {
		return nil, err
	}

	runcExitStatus := make([]byte, 1)
	fd3r.Read(runcExitStatus)

	defer func() {
		theErr = processLogs(log, logFile, theErr)
	}()

	if runcExitStatus[0] != 0 {
		return nil, fmt.Errorf("exit status %d", runcExitStatus[0])
	}

	return d.newProcess(cmd), nil
}

func (d *ExecRunner) Attach(log lager.Logger, processID string, io garden.ProcessIO, processesPath string) (garden.Process, error) {
	return d.iodaemonRunner.Attach(log, processID, io, processesPath)
}

type process struct {
	wait func() error
}

func (d *ExecRunner) newProcess(cmd *exec.Cmd) *process {
	exitCh := make(chan struct{})
	var exitErr error
	go func() {
		exitErr = d.commandRunner.Wait(cmd)
		close(exitCh)
	}()

	return &process{
		wait: func() error {
			<-exitCh
			return exitErr
		},
	}
}

func (p *process) ID() string {
	return ""
}

func (p *process) Wait() (int, error) {
	if err := p.wait(); err != nil {
		exitError, ok := err.(ExitError)
		if !ok {
			return 255, err
		}

		waitStatus, ok := exitError.Sys().(ExitStatuser)
		if !ok {
			return 255, err
		}

		return waitStatus.ExitStatus(), nil
	}

	return 0, nil
}

func (p *process) SetTTY(garden.TTYSpec) error {
	return nil
}

func (p *process) Signal(garden.Signal) error {
	return nil
}

type ExitError interface {
	Sys() interface{}
}

type ExitStatuser interface {
	ExitStatus() int
}

type fifos [3]struct {
	Name     string
	Path     string
	CopyTo   io.Writer
	CopyFrom io.Reader
	Open     func(p string) (*os.File, error)
}

func mkFifos(pio garden.ProcessIO, stdin, stdout, stderr string) (fifos, []string, error) {
	pipes := fifos{
		{Name: "stdin", Path: stdin, CopyFrom: pio.Stdin, Open: func(p string) (*os.File, error) { return os.OpenFile(p, os.O_WRONLY, 0600) }},
		{Name: "stdout", Path: stdout, CopyTo: pio.Stdout, Open: os.Open},
		{Name: "stderr", Path: stderr, CopyTo: pio.Stderr, Open: os.Open},
	}

	pipeArgs := []string{}
	for _, pipe := range pipes {
		pipeArgs = append(pipeArgs, fmt.Sprintf("-%s", pipe.Name), pipe.Path)
		if err := syscall.Mkfifo(pipe.Path, 0); err != nil {
			return pipes, nil, err
		}
	}

	return pipes, pipeArgs, nil
}

func (f fifos) start() error {
	for _, pipe := range f {
		r, err := pipe.Open(pipe.Path)
		if err != nil {
			return err
		}

		if pipe.CopyFrom != nil {
			go io.Copy(r, pipe.CopyFrom)
		}

		if pipe.CopyTo != nil {
			go io.Copy(pipe.CopyTo, r)
		}
	}

	return nil
}

func contains(envVars []string, envVar string) bool {
	for _, e := range envVars {
		if e == envVar {
			return true
		}
	}
	return false
}

func processLogs(log lager.Logger, logFile string, err error) error {
	logFileR, openErr := os.Open(logFile)
	if openErr != nil {
		return fmt.Errorf("start: read log file: %s", openErr)
	}

	buff, readErr := ioutil.ReadAll(logFileR)
	if readErr != nil {
		return fmt.Errorf("start: read log file: %s", readErr)
	}

	forwardRuncLogsToLager(log, buff)

	if err != nil {
		return wrapWithErrorFromRuncLog(log, err, buff)
	}

	return nil
}

func forwardRuncLogsToLager(log lager.Logger, buff []byte) {
	parsedLogLine := struct{ Msg string }{}
	for _, logLine := range strings.Split(string(buff), "\n") {
		if err := logfmt.Unmarshal([]byte(logLine), &parsedLogLine); err == nil {
			log.Debug("runc", lager.Data{
				"message": parsedLogLine.Msg,
			})
		}
	}
}

func wrapWithErrorFromRuncLog(log lager.Logger, originalError error, buff []byte) error {
	parsedLogLine := struct{ Msg string }{}
	logfmt.Unmarshal(buff, &parsedLogLine)
	return fmt.Errorf("runc exec: %s: %s", originalError, parsedLogLine.Msg)
}
