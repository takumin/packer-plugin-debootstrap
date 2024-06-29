//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package debootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/chroot"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

const BuilderId = "debootstrap.builder"

type wrappedCommand struct {
	Command string
}

type Builder struct {
	config Config
	runner multistep.Runner
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Suite     string `mapstructure:"suite" required:"true"`
	MirrorURL string `mapstructure:"mirror_url" required:"true"`

	MountPath         string     `mapstructure:"mount_path" required:"false"`
	MountDevice       string     `mapstructure:"mount_device" required:"false"`
	MountChrootDevice [][]string `mapstructure:"mount_chroot_device" required:"false"`

	CommandWrapper string `mapstructure:"command_wrapper" required:"false"`

	ctx interpolate.Context
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec {
	return b.config.FlatMapstructure().HCL2Spec()
}

func (b *Builder) Prepare(raws ...interface{}) (generatedVars []string, warnings []string, err error) {
	err = config.Decode(&b.config, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"command_wrapper",
			},
		},
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	mirrorURL, ok := os.LookupEnv("DEBOOTSTRAP_MIRROR_URL")
	if ok {
		mirrorURL = strings.TrimSpace(mirrorURL)
		if mirrorURL != "" {
			b.config.MirrorURL = mirrorURL
		}
	}

	if b.config.CommandWrapper == "" {
		b.config.CommandWrapper = "sudo {{.Command}}"
	}

	if b.config.MountPath == "" {
		b.config.MountPath = path.Join(os.TempDir(), "rootfs")
	}

	if b.config.MountDevice == "" {
		b.config.MountDevice = "tmpfs"
	}

	if b.config.MountChrootDevice == nil {
		b.config.MountChrootDevice = make([][]string, 0)
	}

	if len(b.config.MountChrootDevice) == 0 {
		b.config.MountChrootDevice = [][]string{
			{"devtmpfs", "devtmpfs", "/dev", ""},
			{"devpts", "devpts", "/dev/pts", "gid=5,mode=620"},
			{"proc", "proc", "/proc", ""},
			{"sysfs", "sysfs", "/sys", ""},
			{"tmpfs", "tmpfs", "/run", "mode=755"},
			{"tmpfs", "tmpfs", "/tmp", ""},
		}
	}

	var errs *packer.MultiError

	if b.config.Suite == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("required suite"))
	}

	if b.config.MirrorURL == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("required mirror_url"))
	}

	mount_path, err := filepath.Abs(b.config.MountPath)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("debootstrap target dir: %w", err))
	}
	b.config.MountPath = mount_path

	mirrorUrl, err := url.Parse(b.config.MirrorURL)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("debootstrap mirror url: %w", err))
	}
	switch mirrorUrl.Scheme {
	case "http":
	case "https":
	default:
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("debootstrap mirror url: unknown scheme: %s", mirrorUrl.Scheme))
	}
	b.config.MirrorURL = mirrorUrl.String()

	switch b.config.MountDevice {
	case "tmpfs":
	default:
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("unsupported mount device type: %s", b.config.MountDevice))
	}

	for _, mounts := range b.config.MountChrootDevice {
		if len(mounts) != 4 {
			errs = packer.MultiErrorAppend(
				errs, errors.New("mount_chroot_device requires 4 elements"))
			break
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return nil, warnings, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	steps := []multistep.Step{
		&StepPrepareMountPath{
			MountPath: b.config.MountPath,
		},
		&StepMountDevice{
			MountPath:   b.config.MountPath,
			MountDevice: b.config.MountDevice,
		},
		&StepDebootstrap{
			Suite:     b.config.Suite,
			MountPath: b.config.MountPath,
			MirrorURL: b.config.MirrorURL,
		},
		&StepMountChrootDevice{
			MountChrootDevice: b.config.MountChrootDevice,
		},
		&chroot.StepChrootProvision{},
	}

	wrappedCommand := func(command string) (string, error) {
		ictx := b.config.ctx
		ictx.Data = &wrappedCommand{Command: command}
		return interpolate.Render(b.config.CommandWrapper, &ictx)
	}

	state := new(multistep.BasicStateBag)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("wrappedCommand", common.CommandWrapper(wrappedCommand))

	b.runner = commonsteps.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(ctx, state)

	if err, ok := state.GetOk("error"); ok {
		return nil, err.(error)
	}

	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("build was cancelled")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("build was halted")
	}

	return &Artifact{}, nil
}
