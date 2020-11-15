package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/joesonw/distress/cmd/agent/app"
)

func main() {

	var logger *zap.Logger
	envs := map[string]string{}

	rootCmd := &cobra.Command{
		Use:   "distress-agent",
		Short: "distributed stress agent",
	}

	pDebug := rootCmd.PersistentFlags().Bool("debug", false, "enable debug mode")
	pName := rootCmd.PersistentFlags().String("name", "", "agent/job name")
	pEnvs := rootCmd.PersistentFlags().StringArrayP("env", "e", nil, "set lua script environment variables")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if *pName == "" {
			return fmt.Errorf("--name is required")
		}

		var err error
		if *pDebug {
			logger, err = zap.NewDevelopment()
		} else {
			logger, err = zap.NewProduction()
		}
		if err != nil {
			return err
		}

		for _, env := range *pEnvs {
			kvs := strings.Split(env, "=")
			if len(kvs) >= 2 {
				envs[kvs[0]] = strings.Join(kvs[1:], "=")
			}
		}

		return nil
	}

	rootCmd.AddCommand(app.MakeCmdRun(&logger, pName, pDebug, envs))

	err := rootCmd.Execute()
	if err != nil {
		println(err)
		os.Exit(1)
	}
}
