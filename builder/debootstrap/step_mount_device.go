package debootstrap

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepMountDevice struct {
	MountPath   string
	MountDevice string
}

func (s *StepMountDevice) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	switch s.MountDevice {
	case "tmpfs":
		mountCommand, err := wrappedCommand(fmt.Sprintf(
			"mount -t tmpfs tmpfs %s",
			s.MountPath,
		))
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		cmd := common.ShellCommand(mountCommand)
		if err := cmd.Run(); err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	default:
		ui.Error(fmt.Sprintf("unsupported mount device type: %s", s.MountDevice))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepMountDevice) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	cleanupCommand, err := wrappedCommand(fmt.Sprintf(
		"umount %s",
		s.MountPath,
	))
	if err != nil {
		ui.Error(err.Error())
		return
	}

	cmd := common.ShellCommand(cleanupCommand)
	if err := cmd.Run(); err != nil {
		ui.Error(err.Error())
		return
	}
}
