package job

import "github.com/spf13/cobra"

func NewJobCmd() *cobra.Command {
	jobCmd := &cobra.Command{
		Use:   "job",
		Short: "Manage scheduled jobs",
	}

	jobCmd.AddCommand(NewJobAddCmd())
	jobCmd.AddCommand(NewJobRemoveCmd())
	jobCmd.AddCommand(NewJobListCmd())
	jobCmd.AddCommand(NewJobUpdateCmd())

	return jobCmd
}
