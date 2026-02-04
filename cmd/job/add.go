package job

import (
	validate "github.com/gclkaze/evamon/cmd/internal"
	"github.com/gclkaze/evamon/cmd/internal/app"
	"github.com/gclkaze/evamon/cmd/internal/models"
	"github.com/spf13/cobra"
)

func NewJobAddCmd(application *app.Evamon) *cobra.Command {
	var widgetPath string
	var argsCSV string
	var tagsCSV string
	var description string
	var allowOverlap bool

	var jobAddCmd = &cobra.Command{
		Use:   "add <interval-string> <eva-script-path>",
		Short: "Add a new job",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, argv []string) {
			schedule := argv[0]
			scriptPath := argv[1]

			argsList := validate.ParseCSV(argsCSV)
			tagsList := validate.ParseCSV(tagsCSV)

			// Construct (preprocess happens inside)
			req := models.NewJobAddRequest(schedule, scriptPath, argsList, description, tagsList, allowOverlap)

			// Validate (all rules inside)
			if err := req.IsValid(); err != nil {
				application.GetPrinter().Error(err)
				return
			}

			// widgetPath is NOT part of request: validate separately
			if widgetPath != "" {
				if err := validate.RequireFileExists(widgetPath); err != nil {
					application.GetPrinter().Error(err)
					return
				}
				absWidget, err := validate.RequireAbsPath(widgetPath)
				if err != nil {
					application.GetPrinter().Error(err)
					return
				}
				widgetPath = absWidget
			}

			if err := application.AddJob(req, widgetPath); err != nil {
				application.GetPrinter().Error(err)
				return
			}

			application.GetPrinter().Info("Job added successfully")
		},
	}

	jobAddCmd.Flags().StringVar(&widgetPath, "widget-path", "", "Optional widget path")
	jobAddCmd.Flags().StringVar(&argsCSV, "args", "", "Optional args list (comma-separated)")
	jobAddCmd.Flags().StringVar(&tagsCSV, "tags", "", "Optional tags list (comma-separated)")
	jobAddCmd.Flags().StringVar(&description, "description", "", "Optional description")
	jobAddCmd.Flags().BoolVar(&allowOverlap, "allow-overlap", false, "Allow overlapping runs (default false)")

	return jobAddCmd
}
