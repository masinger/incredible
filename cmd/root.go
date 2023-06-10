package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/masinger/incredible/internal/interactive/pterm"
	zap2 "github.com/masinger/incredible/internal/interactive/zap"
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
var nonInteractive bool
var executionOptions = execution.Options{}

var rootCmd = &cobra.Command{
	Use: "incredible <CMD>",
	Example: `# The following will run a new instance of bash, having access to all mapped environment variables:
incredible bash

#Similar, the following will run kubectl with a KUBECONFIG sourced from a secret source:
incredible kubectl get pod `,
	Short: "Runs the provided command with environment variables and temporary files sourced from a safe source",
	Args:  cobra.MinimumNArgs(1),
	Long: `Incredible is a helper tool, that helps users to obtain secret values from a safe source
(e.g. password managers like Bitwarden, Cloud Stores like Azure Key Vaults and others).

A common way to provide access tokens, passwords, certificates 
and other secrets, is by using an environment variable holding the required
secret itself or pointing to a file containing it.`,
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

		if nonInteractive {
			logging.Interactive = zap2.NewZapInteractive(logging.Logger.Desugar())
		} else {
			logging.Interactive = pterm.NewPtermInteractive()
		}

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

		action := logging.Interactive.StartAction("Incredible is loading your environment")

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

		action.Complete("Done loading your environment.")
		return childCmd.Run()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "If present, debug output is enabled.")
	rootCmd.PersistentFlags().BoolVar(&nonInteractive, "non-interactive", false, "If present, interactive mode is disabled.")
}
