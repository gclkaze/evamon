package view

import (
	"github.com/gclkaze/evamon/cmd/internal/app"
	"github.com/spf13/cobra"
)

func NewViewCmd(app *app.Evamon) *cobra.Command {
	viewCmd := &cobra.Command{
		Use:   "view",
		Short: "Manage views",
	}

	viewCmd.AddCommand(NewViewAttachCmd(app))
	viewCmd.AddCommand(NewViewLsCmd(app))
	viewCmd.AddCommand(NewViewRenderCmd(app))
	viewCmd.AddCommand(NewDashboardRenderCmd(app))
	

	return viewCmd
}
