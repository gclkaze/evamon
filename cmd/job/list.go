package job

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewJobListCmd() *cobra.Command {
	var jobListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List jobs",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			// No validations needed (no args)
			// TODO: send over websocket
			fmt.Println("Listing jobs...")
		},
	}
	return jobListCmd
}
