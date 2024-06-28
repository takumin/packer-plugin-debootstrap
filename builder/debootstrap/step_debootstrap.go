package debootstrap

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepDebootstrap struct {
	suite      string
	mount_path string
	mirrorURL  string
}

func (s *StepDebootstrap) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	debootstrapCommand, err := wrappedCommand(fmt.Sprintf(
		"debootstrap %s %s %s",
		s.suite,
		s.mount_path,
		s.mirrorURL,
	))
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// #nosec G204
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", debootstrapCommand)
	ui.Message(fmt.Sprintf("Deboostrap command: %v", debootstrapCommand))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		ui.Error(err.Error())
		s.Cleanup(state)
		return multistep.ActionHalt
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		ui.Error(err.Error())
		s.Cleanup(state)
		return multistep.ActionHalt
	}

	if err := cmd.Start(); err != nil {
		ui.Error(err.Error())
		s.Cleanup(state)
		return multistep.ActionHalt
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			ui.Message(scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			ui.Error(scanner.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		ui.Error(err.Error())
		s.Cleanup(state)
		return multistep.ActionHalt
	}

	state.Put("mount_path", s.targetDir)

	return multistep.ActionContinue
}

func (s *StepDebootstrap) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	cleanupCommand, err := wrappedCommand(fmt.Sprintf(
		"rm -fr %s",
		s.targetDir,
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
