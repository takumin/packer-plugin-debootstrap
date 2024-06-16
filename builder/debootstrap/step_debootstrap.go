package debootstrap

import (
	"bufio"
	"context"
	"os/exec"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepDebootstrap struct {
	suite     string
	targetDir string
	mirrorURL string
}

func (s *StepDebootstrap) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	// #nosec G204
	cmd := exec.CommandContext(ctx, "sudo", "debootstrap", s.suite, s.targetDir, s.mirrorURL)

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
}
