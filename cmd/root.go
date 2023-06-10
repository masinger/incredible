package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/masinger/incredible/pkg/execution"
	"github.com/masinger/incredible/pkg/logging"
	"github.com/masinger/incredible/pkg/provider"
	"github.com/masinger/incredible/pkg/specs/loader"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	exec2 "os/exec"
)

var debug bool
var executionOptions = execution.Options{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "incredible",
	Short: "A brief description of your application",
	Args:  cobra.MinimumNArgs(1),
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var logger *zap.Logger
		var err error
		if debug {
			logger, err = zap.NewDevelopment()
		} else {
			logger, err = zap.NewProduction()
		}
		if err != nil {
			return err
		}
		logging.Logger = logger.Sugar()
		executionOptions.Log = logging.Logger
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		ctx := context.TODO()

		manifest, err := loader.DefaultLoader()
		if err != nil {
			return err
		}
		if manifest == nil {
			return fmt.Errorf("could not find incredible manifest")
		}

		exec, err := execution.NewExecution(
			provider.Providers,
			executionOptions,
		)
		if err != nil {
			return err
		}

		customizer, err := exec.LoadSources(ctx, manifest)
		if err != nil {
			return err
		}
		childCmd := exec2.CommandContext(ctx, args[0], args[1:]...)
		childCmd.Stdout = os.Stdout
		childCmd.Stdin = os.Stdin
		childCmd.Stderr = os.Stderr
		childCmd.Env = childCmd.Environ()

		cleanup, err := customizer(childCmd)
		if cleanup != nil {
			defer func() {
				if cleanupErr := cleanup(childCmd); cleanupErr != nil {
					err = errors.Join(err, cleanupErr)
				}
			}()
		}
		if err != nil {
			return err
		}

		return childCmd.Run()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.incredible.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Set this flag to enable debug output.")
}
