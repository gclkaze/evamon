package view

import (
	"github.com/gclkaze/evamon/cmd/internal/app"
	"github.com/gclkaze/evamon/cmd/internal/userinput"
	"github.com/spf13/cobra"
)

func NewDashboardRenderCmd(application *app.Evamon) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Render a stored Dashboard view project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			params := userinput.NewViewDashboardRenderParams(path)
			err := params.IsValid()
			if err != nil {
				application.GetPrinter().Error(err)
				return
			}

			if err := application.RenderDashboardViewProject(params); err != nil {
				application.GetPrinter().Error(err)
				return
			}
		},
	}

	return cmd
}
