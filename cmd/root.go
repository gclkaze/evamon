package cmd

import (
	"fmt"
	"os"

	"github.com/gclkaze/evamon/cmd/job"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "evamon",
		Short: "evamon is a CLI for managing Evacron jobs",
	}

	// Add child commands
	rootCmd.AddCommand(job.NewJobCmd())

	return rootCmd
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		// Cobra-level parsing errors still come here (unknown command, bad args, etc).
		// You said you'll handle errors yourself — this is only for Cobra execution errors.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
