package runrunc

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/gunk/command_runner"
)

type Creator struct {
	runcPath      string
	commandRunner command_runner.CommandRunner
}

func NewCreator(runcPath string, commandRunner command_runner.CommandRunner) *Creator {
	return &Creator{
		runcPath, commandRunner,
	}
}

func (c *Creator) Create(log lager.Logger, bundlePath, id string, _ garden.ProcessIO) (theErr error) {
	log = log.Session("create", lager.Data{"bundle": bundlePath})

	defer log.Info("finished")

	// HACK?: ensure bundlePath is writable by non-root user
	s := fmt.Sprintf("BUNDLE PATH IS: %s and DIR is: %s", bundlePath, filepath.Dir(bundlePath))
	log.Info(s)
	// BUNDLE PATH IS: /tmp/test-garden-3/containers/eddfd82f-5a18-41a7-6726-799d84ba01cb
	if err := os.Chown(bundlePath, 1000, 1000); err != nil {
		log.Error("chown-failed", err, lager.Data{"path": bundlePath})
		return err
	}

	logFilePath := filepath.Join(bundlePath, "create.log")
	pidFilePath := filepath.Join(bundlePath, "pidfile")

	cmd := exec.Command(c.runcPath, "--debug", "--log", logFilePath, "create", "--bundle", bundlePath, "--pid-file", pidFilePath, id)

	_, stdinW, err := os.Pipe()
	stdoutR, _, err := os.Pipe()
	stderrR, _, err := os.Pipe()

	chownFile := func(file *os.File) error {
		if err := file.Chown(1000, 1000); err != nil {
			log.Error("chown-failed", err, lager.Data{"path": bundlePath})
			return err
		}
		return nil
	}

	chownFile(stdinW)
	chownFile(stdoutR)
	chownFile(stderrR)

	cmd.Stdin = stdinW
	cmd.Stdout = stdoutR
	cmd.Stderr = stderrR

	// HACK: run runc as non-root user
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: 1000,
			Gid: 1000,
		},
	}

	log.Info("creating", lager.Data{
		"runc":        c.runcPath,
		"bundlePath":  bundlePath,
		"id":          id,
		"logPath":     logFilePath,
		"pidFilePath": pidFilePath,
	})

	err = c.commandRunner.Run(cmd)

	defer func() {
		theErr = processLogs(log, logFilePath, err)
	}()

	return
}

func processLogs(log lager.Logger, logFilePath string, upstreamErr error) error {
	logReader, err := os.OpenFile(logFilePath, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("runc create: open log file '%s': %s", logFilePath, err)
	}

	buff, readErr := ioutil.ReadAll(logReader)
	if readErr != nil {
		return fmt.Errorf("runc create: read log file: %s", readErr)
	}

	forwardRuncLogsToLager(log, buff)

	if upstreamErr != nil {
		return wrapWithErrorFromRuncLog(log, upstreamErr, buff)
	}

	return nil
}
