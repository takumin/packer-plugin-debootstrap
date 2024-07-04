package debootstrap

import (
	"fmt"
	"os"
)

type Artifact struct {
	path      string
	StateData map[string]interface{}
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return []string{a.path}
}

func (*Artifact) Id() string {
	return ""
}

func (a *Artifact) String() string {
	return fmt.Sprintf("Rootfs Tarball File: %s", a.path)
}

func (a *Artifact) State(name string) interface{} {
	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	if a.path != "" {
		return os.Remove(a.path)
	}
	return nil
}
