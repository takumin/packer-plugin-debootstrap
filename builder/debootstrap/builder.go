//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package debootstrap

import (
	"context"
	"errors"
	"os"

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
	TargetDir string `mapstructure:"target_dir" required:"true"`
	MirrorURL string `mapstructure:"mirror_url" required:"true"`

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
		b.config.MirrorURL = mirrorURL
	}

	if b.config.CommandWrapper == "" {
		b.config.CommandWrapper = "sudo {{.Command}}"
	}

	var errs *packer.MultiError

	if b.config.Suite == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("required suite"))
	}

	if b.config.TargetDir == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("required target_dir"))
	}

	if b.config.MirrorURL == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("required mirror_url"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return nil, warnings, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	steps := []multistep.Step{
		&StepDebootstrap{
			suite:     b.config.Suite,
			targetDir: b.config.TargetDir,
			mirrorURL: b.config.MirrorURL,
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
