package cmd

import (
	"os"

	"github.com/gclkaze/evamon/cmd/internal/app"
	"github.com/gclkaze/evamon/cmd/internal/services"
	fynediagrams "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/fyne"
	ui "github.com/gclkaze/evamon/cmd/internal/ui/factory"
	"github.com/gclkaze/evamon/cmd/job"
	"github.com/gclkaze/evamon/cmd/view"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	verbose  bool
	hostname string
	port     int
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "evamon",
		Short: "evamon is a CLI for managing Evacron jobs",
	}

	rootCmd.PersistentFlags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"Enable verbose output",
	)

	rootCmd.PersistentFlags().StringVar(
		&hostname,
		"hostname",
		"",
		"Evacron server hostname",
	)

	rootCmd.PersistentFlags().IntVar(
		&port,
		"port",
		0,
		"Evacron server port",
	)

	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("server.hostname", rootCmd.PersistentFlags().Lookup("hostname"))
	_ = viper.BindPFlag("server.port", rootCmd.PersistentFlags().Lookup("port"))

	jobService := services.NewJobService()

	r, err := ui.NewRenderer(ui.KindFyne)
	if err != nil {
		os.Exit(1)
	}
	df := fynediagrams.NewFactory()
	//ws := widgets.NewWidgetService(r, df)
	//_ = ws.CreateWindow(vp)

	//r.Run()

	widgetService := services.NewWidgetService(r, df)
	viewService := services.NewViewService(widgetService)

	app := app.NewEvamon("evamon", verbose, jobService, viewService)
	err = app.Init()
	if err != nil {
		app.GetPrinter().Error(err)
		return nil
	}
	jobService.SetSetup(app)
	viewService.SetSetup(app)

	registryService, err := services.NewProjectsRegistry(app)
	if err != nil {
		app.GetPrinter().Error(err)
		return nil
	}

	app.SetProjectRegistryService(registryService)

	viewService.SetProjectRegistryService(registryService)

	rootCmd.AddCommand(job.NewJobCmd(app))
	rootCmd.AddCommand(view.NewViewCmd(app))

	return rootCmd
}

func Execute() {
	root := NewRootCmd()
	if root == nil {
		os.Exit(1)
	}
	if err := root.Execute(); err != nil {
		// Cobra-level parsing errors still come here (unknown command, bad args, etc).
		// You said you'll handle errors yourself — this is only for Cobra execution errors.
		//fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
