package app

import (
	"fmt"
	"path/filepath"

	"github.com/gclkaze/evamon/cmd/internal/auth"
	"github.com/gclkaze/evamon/cmd/internal/config"
	"github.com/gclkaze/evamon/cmd/internal/fs"
	"github.com/gclkaze/evamon/cmd/internal/models"
	"github.com/gclkaze/evamon/cmd/internal/output"
	"github.com/gclkaze/evamon/cmd/internal/services"
	"github.com/gclkaze/evamon/cmd/internal/wsclient"
	"github.com/gclkaze/evamon/pkg/utils"
	"github.com/magiconair/properties"
)

type Evamon struct {
	appName        string
	logger         output.Printer
	paths          *fs.Paths
	properties     *properties.Properties
	propsPath      string
	widgetFilePath string

	jobService *services.JobService
	resolver   *config.EndpointResolver
	ep         *config.Endpoint
	client     *wsclient.Client
	token      string
}

func NewEvamon(appName string, verbose bool, jobService *services.JobService) *Evamon {

	return &Evamon{appName: appName, logger: output.NewConsolePrinter(verbose), jobService: jobService}
}

func (inst Evamon) GetEndpointResolver() *config.EndpointResolver {
	return inst.resolver
}

func (inst Evamon) GetToken() string {
	return inst.token
}
func (inst Evamon) GetProperties() *properties.Properties {
	return inst.properties
}

func (inst Evamon) GetPrinter() output.Printer {
	return inst.logger
}

func (inst Evamon) GetWSClient() *wsclient.Client {
	if inst.client != nil {
		return inst.client
	}
	inst.client = wsclient.NewWSClient(inst.ep.WSURL(), inst.token, inst.logger)
	return inst.client
	/*	if err := client.Connect(context.Background()); err != nil {
		return
	}*/
}
func (inst *Evamon) Init() error {
	paths, propsPath, exists, err := fs.EnsurePropertiesFile(inst.appName, "application.properties")
	if err != nil {
		return err
	}

	inst.paths = &paths
	inst.propsPath = propsPath

	inst.logger.Info("Config dir:" + paths.Config)
	inst.logger.Info("Lib dir:" + paths.Lib)
	inst.logger.Info("Logs dir:" + paths.Logs)
	inst.logger.Info("Properties:" + propsPath + "exists? " + utils.BoolToString(exists))

	if !exists {
		defaultProps := []byte("port=8123\nlog.level=info\njobs_file=jobs.json")
		// safest write:
		if err = fs.WriteFileAtomic(propsPath, defaultProps, 0o644); err != nil {
			inst.logger.ErrorWithMessage("create default properties failed:", err)
			return err
		}
		inst.logger.Info("Created default properties:" + propsPath)
	}

	err = inst.readProperties()
	if err != nil {
		return err
	}

	inst.resolver = config.NewEndpointResolver(inst.properties).WithKeys("server.hostname", "server.port", "hostname", "port")
	ep, err := inst.resolver.Resolve()
	if err != nil {
		inst.logger.Error(err)
		return nil
	}

	inst.ep = &ep

	token, err := auth.NewTokenLoader(inst.properties, "auth.token_file").LoadToken()
	if err != nil {
		inst.logger.Error(fmt.Errorf("couldn't load token file"))
		return nil
	}
	inst.token = token

	err = inst.setupFolders()
	if err != nil {
		return err
	}
	return nil
}
func (inst *Evamon) readProperties() error {
	inst.properties = properties.MustLoadFile(inst.propsPath, properties.UTF8)
	if inst.properties == nil {
		return fmt.Errorf("couldn't read application properties file %s", inst.propsPath)
	}
	return nil
}

func (inst *Evamon) setupFolders() error {
	if inst.properties == nil {
		return fmt.Errorf("couldn't read properties")
	}

	if inst.paths == nil {
		return fmt.Errorf("couldn't setup the app folders")
	}

	fn := inst.properties.GetString("widget_file", "job.json")

	thePaths := []string{inst.paths.Config, inst.paths.Lib, inst.paths.Logs}
	for i := range thePaths {
		if !utils.FolderExists(thePaths[i]) {
			err := utils.CreateFolder(thePaths[i])
			if err != nil {
				return err
			}
		}
	}

	widgetFile := filepath.Join(inst.paths.Lib, fn)
	if !utils.FileExists(widgetFile) {
		if err := fs.WriteFileAtomic(widgetFile, []byte("{}"), 0o644); err != nil {
			inst.logger.ErrorWithMessage("create default properties failed:", err)
			return err
		}
	}

	inst.widgetFilePath = widgetFile

	return nil
}

func (inst *Evamon) AddJob(job *models.JobAddRequest, widgetPath string) error {
	return inst.jobService.AddJob(job, widgetPath)
}
