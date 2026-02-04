package cmd

import (
	"os"

	"github.com/gclkaze/evamon/cmd/internal/app"
	"github.com/gclkaze/evamon/cmd/internal/services"
	"github.com/gclkaze/evamon/cmd/job"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	verbose  bool
	hostname string
	port     int
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "evamon",
		Short: "evamon is a CLI for managing Evacron jobs",
	}

	rootCmd.PersistentFlags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"Enable verbose output",
	)

	rootCmd.PersistentFlags().StringVar(
		&hostname,
		"hostname",
		"",
		"Evacron server hostname",
	)

	rootCmd.PersistentFlags().IntVar(
		&port,
		"port",
		0,
		"Evacron server port",
	)

	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("server.hostname", rootCmd.PersistentFlags().Lookup("hostname"))
	_ = viper.BindPFlag("server.port", rootCmd.PersistentFlags().Lookup("port"))

	jobService := services.NewJobService()
	app := app.NewEvamon("evamon", verbose, jobService)
	err := app.Init()
	if err != nil {
		app.GetPrinter().Error(err)
		return nil
	}
	jobService.SetSetup(app)

	rootCmd.AddCommand(job.NewJobCmd(app))

	return rootCmd
}

func Execute() {
	root := NewRootCmd()
	if root == nil {
		os.Exit(1)
	}
	if err := root.Execute(); err != nil {
		// Cobra-level parsing errors still come here (unknown command, bad args, etc).
		// You said you'll handle errors yourself — this is only for Cobra execution errors.
		//fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
