package job

import (
	"fmt"
	"os"

	validate "github.com/gclkaze/evamon/cmd/internal"
	"github.com/spf13/cobra"
)

func NewJobUpdateCmd() *cobra.Command {
	var widgetPath string

	var jobUpdateCmd = &cobra.Command{
		Use:   "update <job-id> <interval-string> <eva-script-path>",
		Short: "Update an existing job",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			jobID := args[0]
			intervalString := args[1]
			evaScriptPath := args[2]

			// Validations
			if err := validate.IsValidId(jobID); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				return
			}
			if err := validate.IntervalString(intervalString); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				return
			}
			if err := validate.RequireFileExists(evaScriptPath); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				return
			}
			if widgetPath != "" {
				if err := validate.RequireFileExists(widgetPath); err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
					return
				}
			}

			// Normalize paths
			absEva, err := validate.RequireAbsPath(evaScriptPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				return
			}
			evaScriptPath = absEva

			if widgetPath != "" {
				absWidget, err := validate.RequireAbsPath(widgetPath)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
					return
				}
				widgetPath = absWidget
			}

			// TODO: send over websocket
			fmt.Printf("Updating job %s:\n  interval-string: %s\n  eva-script-path: %s\n", jobID, intervalString, evaScriptPath)
			if widgetPath != "" {
				fmt.Printf("  widget-path: %s\n", widgetPath)
			}
		},
	}

	jobUpdateCmd.Flags().StringVar(&widgetPath, "widget-path", "", "Optional widget path")
	return jobUpdateCmd
}
