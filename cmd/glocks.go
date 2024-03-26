package main

import (
	"errors"
	"github.com/levpaul/glocks/internal/interpreter"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
)

func main() {
	execute()
}

func execute() {
	logCfg := zap.NewDevelopmentConfig()
	logCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	logCfg.DisableStacktrace = true
	logCfg.DisableCaller = true
	logCfg.EncoderConfig.TimeKey = "" // Disable timestamps

	rawLogger, _ := logCfg.Build()
	defer rawLogger.Sync() // flushes buffer, if any
	log := rawLogger.Sugar()

	glocksI := interpreter.New(log)
	var rootCmd = &cobra.Command{
		Use:           "glocks",
		Short:         "glocks <file> run <file> or open the glocks REPL",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				log.Error("Expected maximum of 1 arg - received '", args, "' - exiting 1")
				return errors.New("too many args")
			}

			if len(args) == 1 {
				program, err := os.ReadFile(args[0])
				if err != nil {
					log.With("error", err).Errorf("Failed to read file '%s' from disk\n", args[0])
					return err
				}
				return glocksI.Run(string(program))
			}

			return glocksI.REPL()
		},
	}

	if err := rootCmd.Execute(); err != nil {
		// Cobra logic is expected to print human friendly error
		log.Debug("error code: ", err.Error())
		os.Exit(1)
	}

}
