package imageplugin

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden-shed/rootfs_provider"
	"code.cloudfoundry.org/lager"
)

type PrivilegedCommandCreator struct {
	BinPath   string
	ExtraArgs []string
}

func (p *PrivilegedCommandCreator) CreateCommand(log lager.Logger, handle string, spec rootfs_provider.Spec) *exec.Cmd {
	args := append(p.ExtraArgs, "create")
	if spec.QuotaSize > 0 {
		args = append(args, "--disk-limit-size-bytes", strconv.FormatInt(spec.QuotaSize, 10))

		if spec.QuotaScope == garden.DiskLimitScopeExclusive {
			args = append(args, "--exclude-image-from-quota")
		}
	}

	rootfs := strings.Replace(spec.RootFS.String(), "#", ":", 1)

	args = append(args, rootfs, handle)
	cmd := exec.Command(p.BinPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: 0,
			Gid: 0,
		},
	}

	return cmd
}

func (p *PrivilegedCommandCreator) DestroyCommand(log lager.Logger, handle string) *exec.Cmd {
	cmd := exec.Command(p.BinPath, append(p.ExtraArgs, "delete", handle)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: 0,
			Gid: 0,
		},
	}

	return cmd
}

func (p *PrivilegedCommandCreator) MetricsCommand(log lager.Logger, handle string) *exec.Cmd {
	cmd := exec.Command(p.BinPath, append(p.ExtraArgs, "stats", handle)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: 0,
			Gid: 0,
		},
	}

	return cmd
}

func (p *PrivilegedCommandCreator) GCCommand(log lager.Logger) *exec.Cmd {
	log = log.Session("image-plugin-gc")
	log.Debug("start")
	defer log.Debug("end")

	// args := append(p.extraArgs, "clean")
	// cmd := exec.Command(p.binPath, args...)
	// cmd.Stderr = lagregator.NewRelogger(log)
	// outBuffer := bytes.NewBuffer([]byte{})
	// cmd.Stdout = outBuffer

	// if err := p.commandRunner.Run(cmd); err != nil {
	// 	logData := lager.Data{"action": "clean", "stdout": outBuffer.String()}
	// 	log.Error("external-image-manager-result", err, logData)
	// 	return fmt.Errorf("external image manager clean failed: %s (%s)", outBuffer.String(), err)
	// }

	return nil
}
