package debootstrap

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepDebootstrap struct {
	Suite     string
	Variant   string
	MountPath string
	MirrorURL string
}

func (s *StepDebootstrap) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	flags := make([]string, 0, 16)

	if s.Variant != "" {
		flags = append(flags, "--variant", s.Variant)
	}

	flags = append(flags, s.Suite)
	flags = append(flags, s.MountPath)
	flags = append(flags, s.MirrorURL)

	debootstrapCommand, err := wrappedCommand(fmt.Sprintf(
		"debootstrap %s",
		strings.Join(flags, " "),
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

	return multistep.ActionContinue
}

func (s *StepDebootstrap) Cleanup(state multistep.StateBag) {
}
