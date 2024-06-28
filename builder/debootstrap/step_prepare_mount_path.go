package debootstrap

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepPrepareMountPath struct {
	MountPath string
}

func (s *StepPrepareMountPath) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	createDirCommand, err := wrappedCommand(fmt.Sprintf(
		"mkdir -p %s",
		s.MountPath,
	))
	if err != nil {
		ui.Error(err.Error())
		s.Cleanup(state)
		return multistep.ActionHalt
	}

	cmd := common.ShellCommand(createDirCommand)
	if err := cmd.Run(); err != nil {
		ui.Error(err.Error())
		s.Cleanup(state)
		return multistep.ActionHalt
	}

	state.Put("mount_path", s.MountPath)

	return multistep.ActionContinue
}

func (s *StepPrepareMountPath) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	cleanupCommand, err := wrappedCommand(fmt.Sprintf(
		"rm -fr %s",
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
