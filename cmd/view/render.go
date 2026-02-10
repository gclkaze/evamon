package view

import (
	"github.com/gclkaze/evamon/cmd/internal/app"
	"github.com/gclkaze/evamon/cmd/internal/userinput"
	"github.com/spf13/cobra"
)

func NewViewRenderCmd(application *app.Evamon) *cobra.Command {
	var projectID string
	var headless bool

	cmd := &cobra.Command{
		Use:   "render",
		Short: "Render a stored view project",
		Run: func(cmd *cobra.Command, args []string) {
			params := userinput.NewViewRenderParams(projectID, headless)

			if err := application.RenderViewProject(params); err != nil {
				application.GetPrinter().Error(err)
				return
			}
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID to render (required)")
	cmd.Flags().BoolVar(&headless, "headless", false, "Render without UI (headless mode)")
	_ = cmd.MarkFlagRequired("project-id")

	return cmd
}
