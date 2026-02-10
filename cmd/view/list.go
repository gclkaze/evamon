package view

import (
	"github.com/gclkaze/evamon/cmd/internal/app"
	"github.com/gclkaze/evamon/cmd/internal/userinput"
	"github.com/spf13/cobra"
)

func NewViewLsCmd(application *app.Evamon) *cobra.Command {
	var showAll bool
	var noTrunc bool
	var format string
	var filters []string

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List stored view projects",
		Run: func(cmd *cobra.Command, args []string) {
			params := userinput.NewViewLsParams(showAll, noTrunc, format, filters)

			if err := application.PrintProjectRegistry(params); err != nil {
				application.GetPrinter().Error(err)
				return
			}
		},
	}

	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Include entries whose project file is missing/invalid")
	cmd.Flags().BoolVar(&noTrunc, "no-trunc", false, "Do not truncate columns")
	cmd.Flags().StringVar(&format, "format", "", "Format output using a Go template (e.g. '{{.ProjectID}}\\t{{.JobID}}')")
	cmd.Flags().StringArrayVarP(&filters, "filter", "f", nil, "Filter output (repeatable), e.g. -f job=job-123 -f status=OK -f type=boolean")

	return cmd
}
