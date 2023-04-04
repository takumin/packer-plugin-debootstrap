//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package debootstrap

import (
	"errors"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Suite     string `mapstructure:"suite" required:"true"`
	TargetDir string `mapstructure:"target_dir" required:"true"`
	MirrorURL string `mapstructure:"mirror_url" required:"true"`

	ctx interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(c, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)
	if err != nil {
		return nil, err
	}

	var errs *packer.MultiError
	warnings := make([]string, 0)

	if c.Suite == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("required suite"))
	}

	if c.TargetDir == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("required target_dir"))
	}

	if c.MirrorURL == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("required mirror_url"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}
