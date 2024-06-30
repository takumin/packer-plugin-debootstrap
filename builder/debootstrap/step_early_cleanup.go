package debootstrap

import (
	"context"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

type Cleanup interface {
	Cleanup(multistep.StateBag)
}

type StepEarlyCleanup struct{}

func (s *StepEarlyCleanup) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	cleanupKeys := []string{
		"mount_chroot_device_cleanup",
	}

	for _, key := range cleanupKeys {
		c := state.Get(key).(Cleanup)
		log.Printf("[DEBUG] running early cleanup func: %s", key)
		c.Cleanup(state)
	}

	return multistep.ActionContinue
}

func (s *StepEarlyCleanup) Cleanup(state multistep.StateBag) {}
