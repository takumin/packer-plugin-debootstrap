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

type StepRootfsArchive struct {
	MountPath            string
	RootfsArchivePath    string
	RootfsArchiveCommand string
	RootfsArchiveUid     string
	RootfsArchiveGid     string
}

func (s *StepRootfsArchive) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	if s.RootfsArchivePath == "" {
		log.Printf("[DEBUG] skip create rootfs archive")
		return multistep.ActionContinue
	}

	flags := make([]string, 0, 16)
	if s.RootfsArchiveCommand != "" {
		flags = append(flags, "-I", s.RootfsArchiveCommand)
	}
	flags = append(flags, "-p")
	flags = append(flags, "--acls")
	flags = append(flags, "--xattrs")
	flags = append(flags, "--one-file-system")
	flags = append(flags, "-cf", s.RootfsArchivePath)
	flags = append(flags, "-C", s.MountPath)

	archiveCommand, err := wrappedCommand(fmt.Sprintf(
		"tar %s .",
		strings.Join(flags, " "),
	))
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Create Rootfs Archive: %s", s.RootfsArchivePath))
	archiveShellCommand := common.ShellCommand(archiveCommand)
	if err := archiveShellCommand.Run(); err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	chownCommand, err := wrappedCommand(fmt.Sprintf(
		"chown %s:%s %s",
		s.RootfsArchiveUid,
		s.RootfsArchiveGid,
		s.RootfsArchivePath,
	))
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("[DEBUG] rootfs archive owner %s:%s", s.RootfsArchiveUid, s.RootfsArchiveGid)
	chownShellCommand := common.ShellCommand(chownCommand)
	if err := chownShellCommand.Run(); err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepRootfsArchive) Cleanup(state multistep.StateBag) {}
