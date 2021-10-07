package plugin

import (
	"context"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"

	"knative.dev/client/pkg/kn/plugin"

	"knative.dev/kn-plugin-func/cmd"
)

func init() {
	plugin.InternalPlugins = append(plugin.InternalPlugins, &funcPlugin{})
}

type funcPlugin struct{}

func (f *funcPlugin) Name() string {
	return "kn-func"
}

func (f *funcPlugin) Execute(args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()

	rootCmd, _ := cmd.NewRootCmd()
	info, _ := debug.ReadBuildInfo()
	for _, dep := range info.Deps {
		if strings.Contains(dep.Path, "knative.dev/kn-plugin-func") {
			cmd.SetMeta("", dep.Version, dep.Sum)
		}
	}
	oldArgs := os.Args
	defer (func() {
		os.Args = oldArgs
	})()
	os.Args = append([]string{"kn-func"}, args...)
	return rootCmd.ExecuteContext(ctx)
}

// Description for function subcommand visible in 'kn --help'
func (f *funcPlugin) Description() (string, error) {
	return "Function plugin", nil
}

func (f *funcPlugin) CommandParts() []string {
	return []string{"func"}
}

// Path is empty because its an internal plugin
func (f *funcPlugin) Path() string {
	return ""
}
