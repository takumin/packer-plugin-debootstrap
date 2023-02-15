package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"

	exampleBuilder "github.com/takumin/packer-plugin-debootstrap/builder/example"
	exampleDataSource "github.com/takumin/packer-plugin-debootstrap/datasource/example"
	examplePostProcessor "github.com/takumin/packer-plugin-debootstrap/post-processor/example"
	exampleProvisioner "github.com/takumin/packer-plugin-debootstrap/provisioner/example"

	"github.com/takumin/packer-plugin-debootstrap/version"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder(plugin.DEFAULT_NAME, new(exampleBuilder.Builder))
	pps.RegisterProvisioner(plugin.DEFAULT_NAME, new(exampleProvisioner.Provisioner))
	pps.RegisterPostProcessor(plugin.DEFAULT_NAME, new(examplePostProcessor.PostProcessor))
	pps.RegisterDatasource(plugin.DEFAULT_NAME, new(exampleDataSource.Datasource))
	pps.SetVersion(version.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
