package ci

import (
	"fmt"
	"io"
)

const (
	MainLayoutPlainText = `
GitHub Workflow Configuration
  Workflow filepath:  %s
  Workflow name:      %s
  Branch:             %s
  Build:              %s
  Runner:             %s
  Registry login:     %s
  Manual dispatch:    %s
  Workflow overwrite: %s
`
	RequireManyPlainText = `
  Required Secrets & Variables:
    %s
    %s
    %s
    %s
    %s
`

	RequireOnePlainText = `  Required secret:    %s
`

	PostExportManyPlainText = `
GitHub Workflow created at: %s

Create the following Secrets & Variables on github.com:
  %s
  %s
  %s
  %s
  %s
`

	PostExportOnePlainText = `
GitHub Workflow created at: %s

Create the following Secret on github.com: %s
`
)

func PrintConfiguration(w io.Writer, conf CIConfig) error {
	if _, err := fmt.Fprintf(w, MainLayoutPlainText,
		conf.OutputPath(),
		conf.WorkflowName(),
		conf.Branch(),
		builder(conf),
		runner(conf),
		enabledOrDisabled(conf.UseRegistryLogin()),
		enabledOrDisabled(conf.UseWorkflowDispatch()),
		enabledOrDisabled(conf.Force()),
	); err != nil {
		return err
	}

	if conf.UseRegistryLogin() {
		if _, err := fmt.Fprintf(w, RequireManyPlainText,
			secretsPrefix(conf.KubeconfigSecret()),
			secretsPrefix(conf.RegistryPassSecret()),
			varsPrefix(conf.RegistryLoginUrlVar()),
			varsPrefix(conf.RegistryUserVar()),
			varsPrefix(conf.RegistryUrlVar()),
		); err != nil {
			return err
		}

		return nil
	}

	if _, err := fmt.Fprintf(w,
		RequireOnePlainText,
		secretsPrefix(conf.KubeconfigSecret()),
	); err != nil {
		return err
	}

	return nil
}

func PrintPostExportMessage(w io.Writer, conf CIConfig) error {
	if conf.UseRegistryLogin() {
		_, err := fmt.Fprintf(w, PostExportManyPlainText,
			conf.OutputPath(),
			secretsPrefix(conf.KubeconfigSecret()),
			secretsPrefix(conf.RegistryPassSecret()),
			varsPrefix(conf.RegistryLoginUrlVar()),
			varsPrefix(conf.RegistryUserVar()),
			varsPrefix(conf.RegistryUrlVar()),
		)
		return err
	}

	_, err := fmt.Fprintf(w, PostExportOnePlainText,
		conf.OutputPath(),
		secretsPrefix(conf.KubeconfigSecret()),
	)
	return err
}

func builder(conf CIConfig) string {
	if conf.UseRemoteBuild() {
		return "remote"
	}
	return "host"
}

func enabledOrDisabled(value bool) string {
	if value {
		return "enabled"
	}
	return "disabled"
}
