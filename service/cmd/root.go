package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.con/stream-ai/huddle/backend/service/server"
)

var (
	DefaultServerPort = "80"

	ErrProcessingOptions = fmt.Errorf("error processing options")

	// Used for flags.
	cfgFile     string
	userLicense string

	rootCmd = &cobra.Command{
		Use:   "cobra-cli",
		Short: "A generator for Cobra based Applications",
		Long: `Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// ctx := context.Background()

			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			// logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			slog.SetDefault(logger)
			slog.Info("app started", "config-file", viper.ConfigFileUsed())

			configFile, err := cmd.Flags().GetString("config")
			if err != nil {
				return errors.Join(ErrProcessingOptions, err)
			}
			if configFile != "" {
				logger.Info("config file used", "file", viper.ConfigFileUsed())
			}

			// addr := fmt.Sprintf(":%s", viper.GetString("port"))
			addr := fmt.Sprintf(":%s", viper.GetString("port"))
			return server.Run(cmd.Context(), logger, addr)
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.huddle-backend.yaml)")
	rootCmd.PersistentFlags().StringP("port", "p", DefaultServerPort, "server port")
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc")
		viper.AddConfigPath("/app")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".huddle-backend")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			cobra.CheckErr(err)
		}
	}
}
