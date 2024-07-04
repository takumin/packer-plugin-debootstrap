package debootstrap

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepMakeSquashfs struct {
	MountPath       string
	SquashfsPath    string
	SquashfsCommand string
	SquashfsFormat  string
	SquashfsUid     string
	SquashfsGid     string
}

func (s *StepMakeSquashfs) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	if s.SquashfsPath == "" {
		log.Printf("[DEBUG] skip create squashfs")
		return multistep.ActionContinue
	}

	flags := make([]string, 0, 16)
	flags = append(flags, s.MountPath)
	flags = append(flags, s.SquashfsPath)
	if s.SquashfsFormat != "" {
		flags = append(flags, "-comp", s.SquashfsFormat)
	}

	archiveCommand, err := wrappedCommand(fmt.Sprintf(
		"%s %s",
		s.SquashfsCommand,
		strings.Join(flags, " "),
	))
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Create Squashfs: %s", s.SquashfsPath))
	archiveShellCommand := common.ShellCommand(archiveCommand)
	if err := archiveShellCommand.Run(); err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	chownCommand, err := wrappedCommand(fmt.Sprintf(
		"chown %s:%s %s",
		s.SquashfsUid,
		s.SquashfsGid,
		s.SquashfsPath,
	))
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("[DEBUG] rootfs archive owner %s:%s", s.SquashfsUid, s.SquashfsGid)
	chownShellCommand := common.ShellCommand(chownCommand)
	if err := chownShellCommand.Run(); err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepMakeSquashfs) Cleanup(state multistep.StateBag) {}
