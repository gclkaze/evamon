package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gclkaze/evamon/cmd/internal/fs"
	"github.com/gclkaze/evamon/cmd/internal/models"
	"github.com/gclkaze/evamon/cmd/internal/output"
	"github.com/gclkaze/evamon/cmd/internal/userinput"
	"github.com/gclkaze/evamon/cmd/internal/viewproject"
	"github.com/gclkaze/evamon/pkg/utils"
	"golang.org/x/sync/errgroup"
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

func (inst *ViewService) JobIDExists(jobID string) (bool, error) {
	client := inst.setup.GetWSClient()
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		return false, err
	}
	defer client.Close()
	raw, _ := json.Marshal(jobID)
	msg := models.WSMessage{
		Type: "job.exists",
		Data: raw,
	}

	if err := client.SendJSON(ctx, msg); err != nil {
		return false, err
	}

	var resp models.WSMessage
	if err := client.ReadJSON(ctx, &resp); err != nil {
		return false, err
	}
	val, err := resp.DataAsBool()
	if err != nil {
		return false, err
	}

	if !val {
		return false, fmt.Errorf("JobID '%s' does not exist in evacron", jobID)
	}
	return true, nil
}

func (inst *ViewService) JobIDsExist(jobIDs []string) (map[string]bool, error) {
	client := inst.setup.GetWSClient()
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		return nil, err
	}
	defer client.Close()
	raw, _ := json.Marshal(jobIDs)
	msg := models.WSMessage{
		Type: "job.exist",
		Data: raw,
	}

	if err := client.SendJSON(ctx, msg); err != nil {
		return nil, err
	}

	var resp models.WSMessage
	if err := client.ReadJSON(ctx, &resp); err != nil {
		return nil, err
	}
	val, err := resp.DataAsBoolMap()
	if err != nil {
		return nil, err
	}

	if val == nil {
		return nil, fmt.Errorf("the returned truth map is empty")
	}
	return val, nil
}

func (inst *ViewService) ViewAttach(params *userinput.ViewAttachParams) error {
	exists, err := inst.JobIDExists(params.JobID)
	if err != nil {
		return err
	}
	if !exists {
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
	_, err := inst.JobIDExists(vp.JobID)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()                                  // optional: if worker ends, close app
		inst.listenForMessages(ctx, vp.JobID, headless) // scheduler, websockets, etc.
	}()

	// Blocks here, but workers keep running
	err = inst.wservice.CreateProjectUI(vp, inst.setup.GetProperties())
	if err != nil {
		return err
	}

	// Stop background work when app is closing
	inst.wservice.SetOnClosed(func() {
		cancel()
	})

	inst.wservice.Run()
	return nil
}

func (inst *ViewService) listenForMessages(ctx context.Context, jobID string, headless bool) error {
	client := inst.setup.GetWSClient()
	if err := client.Connect(ctx); err != nil {
		return err
	}
	defer client.Close()
	raw, _ := json.Marshal(jobID)
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

	for {
		var resp models.WSMessage
		if err := client.ReadJSONForever(ctx, &resp); err != nil {
			return err
		}

		switch resp.Type {
		case "job.event":
			var e models.ScriptEvent
			if err := json.Unmarshal(resp.Data, &e); err != nil {
				//inst.logger.Error(err)
				continue
			}

			var ex models.ExecutionMessage
			if err := json.Unmarshal([]byte(e.Text), &ex); err != nil {
				//inst.logger.Error(err)
				continue
			}
			inst.dispatch(e.JobID, &ex)

		case "job.listen":
			if resp.Error != "" {
				return fmt.Errorf("listen error: %s", resp.Error)
			}

		default:
		}
	}
}

func (inst *ViewService) RenderDashboard(dp *viewproject.DashboardProject) error {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()                           // optional: if worker ends, close app
		inst.listenForDashboardMessages(ctx, dp) // scheduler, websockets, etc.
	}()

	// Blocks here, but workers keep running
	err := inst.wservice.CreateDashboardProjectUI(dp, inst.setup.GetProperties())
	if err != nil {
		return err
	}

	// Stop background work when app is closing
	inst.wservice.SetOnClosed(func() {
		cancel()
	})

	inst.wservice.Run()
	return nil
}

func (inst *ViewService) listenForDashboardMessages(ctx context.Context, vp *viewproject.DashboardProject) error {
	jobIDs, err := vp.GetUniqueJobIDs()
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	for i := range jobIDs {
		jobID := jobIDs[i]

		g.Go(func() error {
			return inst.listenForMessages(ctx, jobID, false)
		})
	}

	return g.Wait()
}

func (inst *ViewService) dispatch(jobID string, ex *models.ExecutionMessage) {
	if !strings.Contains(ex.Msg, "MetricsIndexSave Operation. Message:") {
		return
	}

	start := strings.Index(ex.Msg, "{")
	if start == -1 {
		return
	}

	var msg models.MetricsMsg
	err := json.Unmarshal([]byte(ex.Msg[start:]), &msg)
	if err != nil {
		return
	}
	inst.wservice.DispatchValue(jobID, msg.Index, time.UnixMilli(ex.T), msg.Value)
	fmt.Print(msg)
}
