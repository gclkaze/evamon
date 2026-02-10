package view

import (
	"github.com/gclkaze/evamon/cmd/internal/app"
	"github.com/gclkaze/evamon/cmd/internal/userinput"
	"github.com/spf13/cobra"
)

func NewViewAttachCmd(application *app.Evamon) *cobra.Command {
	var jobID string
	var widgetPath string

	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach a widget JSON view to a specific job run",
		Run: func(cmd *cobra.Command, args []string) {

			params, err := userinput.NewViewAttachParams(jobID, widgetPath)
			if err != nil {
				application.GetPrinter().Error(err)
				return
			}
			if err := params.IsValid(); err != nil {
				application.GetPrinter().Error(err)
				return
			}

			err = application.ViewAttach(params)
			if err != nil {
				application.GetPrinter().Error(err)
				return
			}
		},
	}

	cmd.Flags().StringVarP(&jobID, "job-id", "j", "", "Job identifier (required)")
	cmd.Flags().StringVarP(&widgetPath, "widget-path", "w", "", "Path to widget JSON file (required)")

	_ = cmd.MarkFlagRequired("job-id")
	_ = cmd.MarkFlagRequired("widget-path")

	return cmd
}
