package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gclkaze/evamon/cmd/internal/fs"
	"github.com/gclkaze/evamon/cmd/internal/models"
	"github.com/gclkaze/evamon/cmd/internal/output"
	"github.com/gclkaze/evamon/cmd/internal/userinput"
	"github.com/gclkaze/evamon/cmd/internal/viewproject"
	"github.com/gclkaze/evamon/pkg/utils"
)

type ViewService struct {
	logger          output.Printer
	setup           MainSetup
	registryService *ProjectsRegistryService
	wservice        *WidgetService
}

func NewViewService(wService *WidgetService) *ViewService {
	return &ViewService{wservice: wService}
}

func (inst *ViewService) SetSetup(setup MainSetup) {
	inst.setup = setup
	inst.logger = setup.GetPrinter()
}

func (inst *ViewService) ViewAttach(params *userinput.ViewAttachParams) error {
	client := inst.setup.GetWSClient()
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		return err
	}
	defer client.Close()
	raw, _ := json.Marshal(params.JobID)
	msg := models.WSMessage{
		Type: "job.exists",
		Data: raw,
	}

	if err := client.SendJSON(ctx, msg); err != nil {
		return err
	}

	var resp models.WSMessage
	if err := client.ReadJSON(ctx, &resp); err != nil {
		return err
	}
	val, err := resp.DataAsBool()
	if err != nil {
		return err
	}

	if !val {
		return fmt.Errorf("JobID '%s' does not exist in evacron", params.JobID)
	}

	viewProject, err := params.ToViewProject()
	if err != nil {
		return err
	}
	viewProject.ID = utils.GetRandomString()
	saveAs := params.GetFileName()
	proj, err := inst.SaveToLib(params.JobID, viewProject, saveAs)
	if err != nil {
		return err
	}
	inst.logger.Info(fmt.Sprintf("View saved at '%s' with ID '%s'", proj, viewProject.ID))
	inst.registryService.Upsert(viewProject.ID, proj)
	return nil
}

func (inst ViewService) SaveToLib(jobID string, vp *viewproject.ViewProject, saveAs string) (string, error) {
	if vp == nil {
		return "", fmt.Errorf("nil view project")
	}

	// Ensure jobID matches the wrapper
	if vp.JobID == "" {
		vp.JobID = jobID
	}
	if vp.JobID != jobID {
		return "", fmt.Errorf("jobID mismatch: arg=%q project=%q", jobID, vp.JobID)
	}

	if err := vp.Validate(); err != nil {
		return "", err
	}

	targetPath := filepath.Join(inst.setup.GetWidgetPath(), jobID)
	if err := os.MkdirAll(targetPath, 0o755); err != nil {
		return "", fmt.Errorf("mkdir %q: %w", filepath.Dir(targetPath), err)
	}
	b, err := json.MarshalIndent(vp, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal project: %w", err)
	}

	targetPath = filepath.Join(targetPath, saveAs)

	if err := fs.WriteFileAtomic(targetPath, b, 0o644); err != nil {
		return "", err
	}

	return targetPath, nil
}
func (inst *ViewService) SetProjectRegistryService(registryService *ProjectsRegistryService) {
	inst.registryService = registryService
}

func (inst *ViewService) Render(vp *viewproject.ViewProject, headless bool) error {
	client := inst.setup.GetWSClient()
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		return err
	}
	defer client.Close()
	raw, _ := json.Marshal(vp.JobID)
	msg := models.WSMessage{
		Type: "job.listen",
		Data: raw,
	}

	if err := client.SendJSON(ctx, msg); err != nil {
		return err
	}

	var ack models.WSMessage
	if err := client.ReadJSON(ctx, &ack); err != nil {
		return err
	}
	if ack.Type != "job.listen" {
		return fmt.Errorf("unexpected response type: %s", ack.Type)
	}
	if ack.Error != "" {
		return fmt.Errorf("listen failed: %s", ack.Error)
	}

	inst.wservice.CreateWindow(vp)

	//	rend, err := ui.NewRenderer(ui.KindFyne)

	for {
		var resp models.WSMessage
		if err := client.ReadJSONForever(ctx, &resp); err != nil {
			return err
		}

		switch resp.Type {
		case "job.event":
			var e models.ScriptEvent
			if err := json.Unmarshal(resp.Data, &e); err != nil {
				inst.logger.Error(err)
				continue
			}
			var s string
			if err := json.Unmarshal(resp.Data, &s); err != nil {
				inst.logger.Error(err)
			} else {
				inst.logger.Info(s)
			}
			inst.logger.Info(fmt.Sprintf("[%s] %s: %s", e.JobID, e.Type, e.Text))

		case "job.listen":
			if resp.Error != "" {
				return fmt.Errorf("listen error: %s", resp.Error)
			}

		default:
		}
	}
	return nil
}
