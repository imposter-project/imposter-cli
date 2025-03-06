package cmd

import (
	"gatehill.io/imposter/internal/engine"
	"github.com/spf13/cobra"
)

var localTypes = []engine.EngineType{
	engine.EngineTypeDockerCore,
	engine.EngineTypeDockerAll,
	engine.EngineTypeDockerDistroless,
	engine.EngineTypeJvmSingleJar,
	engine.EngineTypeGolang,
}

func registerEngineTypeCompletions(cmd *cobra.Command, additionalTypes ...engine.EngineType) {
	_ = cmd.RegisterFlagCompletionFunc("engine-type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var types []string
		for _, t := range localTypes {
			types = append(types, string(t))
		}
		if len(additionalTypes) > 0 {
			for _, t := range additionalTypes {
				types = append(types, string(t))
			}
		}
		return types, cobra.ShellCompDirectiveNoFileComp
	})
}
