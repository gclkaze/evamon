package job

import (
	"github.com/gclkaze/evamon/cmd/internal/app"
	"github.com/spf13/cobra"
)

func NewJobCmd(app *app.Evamon) *cobra.Command {
	jobCmd := &cobra.Command{
		Use:   "job",
		Short: "Manage scheduled jobs",
	}

	jobCmd.AddCommand(NewJobAddCmd(app))
	jobCmd.AddCommand(NewJobRemoveCmd())
	jobCmd.AddCommand(NewJobListCmd())
	jobCmd.AddCommand(NewJobUpdateCmd())

	return jobCmd
}
