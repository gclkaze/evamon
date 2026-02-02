package job

import (
	"fmt"
	"os"

	validate "github.com/gclkaze/evamon/cmd/internal"
	"github.com/spf13/cobra"
)

func NewJobRemoveCmd() *cobra.Command {
	var jobRemoveCmd = &cobra.Command{
		Use:     "remove <job-id>",
		Aliases: []string{"rm"},
		Short:   "Remove a job",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			jobID := args[0]

			if err := validate.IsValidId(jobID); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				return
			}

			// TODO: send over websocket
			fmt.Printf("Removing job: %s\n", jobID)
		},
	}
	return jobRemoveCmd
}
