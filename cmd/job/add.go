package job

import (
	"fmt"
	"os"

	validate "github.com/gclkaze/evamon/cmd/internal"
	"github.com/spf13/cobra"
)

func NewJobAddCmd() *cobra.Command {
	var widgetPath string

	var jobAddCmd = &cobra.Command{
		Use:   "add <interval-string> <eva-script-path>",
		Short: "Add a new job",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			intervalString := args[0]
			evaScriptPath := args[1]

			// Validations
			if err := validate.IntervalString(intervalString); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				return
			}

			// If your eva script must be a file (not a directory), use RequireFileExists
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

			// Optional: normalize to absolute paths
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
			fmt.Printf("Adding job:\n  interval-string: %s\n  eva-script-path: %s\n", intervalString, evaScriptPath)
			if widgetPath != "" {
				fmt.Printf("  widget-path: %s\n", widgetPath)
			}
		},
	}

	jobAddCmd.Flags().StringVar(&widgetPath, "widget-path", "", "Optional widget path")
	return jobAddCmd
}
