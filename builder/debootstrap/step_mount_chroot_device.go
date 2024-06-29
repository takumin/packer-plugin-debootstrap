package debootstrap

import (
	"context"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/moby/sys/mountinfo"
)

type StepMountChrootDevice struct {
	MountChrootDevice [][]string
}

func (s *StepMountChrootDevice) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	mountPath := state.Get("mount_path").(string)
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	for _, mountInfo := range s.MountChrootDevice {
		mountType := mountInfo[0]
		mountDevice := mountInfo[1]
		mountPoint := path.Join(mountPath, mountInfo[2])
		mountOpts := mountInfo[3]

		mkdirCommand, err := wrappedCommand(fmt.Sprintf(
			"mkdir -p %s",
			mountPoint,
		))
		if err != nil {
			ui.Error(err.Error())
			s.Cleanup(state)
			return multistep.ActionHalt
		}

		mkdirShellCommand := common.ShellCommand(mkdirCommand)
		if err := mkdirShellCommand.Run(); err != nil {
			ui.Error(err.Error())
			s.Cleanup(state)
			return multistep.ActionHalt
		}

		mountFlags := make([]string, 0, 8)
		mountFlags = append(mountFlags, "-t", mountType)
		if mountOpts != "" {
			mountFlags = append(mountFlags, "-o", mountOpts)
		}
		mountFlags = append(mountFlags, mountDevice, mountPoint)

		mountCommand, err := wrappedCommand(fmt.Sprintf(
			"mount %s",
			strings.Join(mountFlags, " "),
		))
		if err != nil {
			ui.Error(err.Error())
			s.Cleanup(state)
			return multistep.ActionHalt
		}

		ui.Message(fmt.Sprintf("mount command: %s", mountCommand))
		mountShellCommand := common.ShellCommand(mountCommand)
		if err := mountShellCommand.Run(); err != nil {
			ui.Error(err.Error())
			s.Cleanup(state)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepMountChrootDevice) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	mountPath := state.Get("mount_path").(string)
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	slices.Reverse(s.MountChrootDevice)

	for _, mountInfo := range s.MountChrootDevice {
		mountPoint := path.Join(mountPath, mountInfo[2])

		mounted, err := mountinfo.Mounted(mountPoint)
		if err != nil {
			ui.Error(err.Error())
			return
		}

		if mounted {
			umountCommand, err := wrappedCommand(fmt.Sprintf(
				"umount %s",
				mountPoint,
			))
			if err != nil {
				ui.Error(err.Error())
				return
			}

			ui.Message(fmt.Sprintf("umount command: %s", umountCommand))
			umountShellCommand := common.ShellCommand(umountCommand)
			if err := umountShellCommand.Run(); err != nil {
				ui.Error(err.Error())
				return
			}
		}
	}
}
