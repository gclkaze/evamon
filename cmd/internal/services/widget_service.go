package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gclkaze/evamon/cmd/internal/models"
	"github.com/gclkaze/evamon/cmd/internal/models/ui"
	"github.com/gclkaze/evamon/cmd/internal/output"
	"github.com/gclkaze/evamon/cmd/internal/ui/data"
	"github.com/gclkaze/evamon/cmd/internal/viewproject"
	"github.com/gclkaze/evamon/cmd/internal/wsclient"
	"github.com/magiconair/properties"

	"github.com/gclkaze/evamon/cmd/internal/ui/port"

	porter "github.com/gclkaze/evamon/cmd/internal/ui/diagrams/port"
)

type WidgetService struct {
	logger output.Printer
	setup  MainSetup

	renderer      port.Renderer
	drawerFactory porter.Factory
	//it holds all view projects -> ref -> map[Variable]-> multiple widgets
	variableContainer *ui.VariableContainer
	jobRouter         *ui.JobRouter

	uiHolder        *ui.ProjectUIHolder
	dashboardHolder *ui.DashboardUIHolder

	triggerSenderFactory func(jobID, diagramID string) data.TriggerSendFunc
	streamTracker        *triggerStreamTracker
}

func NewWidgetService(r port.Renderer, df porter.Factory) *WidgetService {
	return &WidgetService{renderer: r, drawerFactory: df, variableContainer: ui.NewVariableContainer(), jobRouter: ui.NewJobRouter()}
}

func (inst *WidgetService) logErr(err error) {
	if inst.logger != nil {
		inst.logger.Error(err)
	}
}

func (inst *WidgetService) SetSetup(setup MainSetup) {
	inst.setup = setup
	inst.logger = setup.GetPrinter()
	inst.streamTracker = newTriggerStreamTracker()

	inst.triggerSenderFactory = func(jobID, diagramID string) data.TriggerSendFunc {
		return func(ruleID string, files []string) {
			go inst.runTriggerProtocol(jobID, ruleID, diagramID, files)
		}
	}
}

const (
	TRIGGER_PROTOCOL_ACKNOWLEDGEMENT = "ACK"
	TRIGGER_STREAM_END               = "STREAM-END"
)

// runTriggerProtocol orchestrates the full trigger-operation protocol.
func (inst *WidgetService) runTriggerProtocol(jobID, ruleID, diagramID string, files []string) {
	msg := models.NewTriggerOperationMsg(jobID, ruleID, diagramID, files)
	wsMsg, err := msg.ToWSMessage()
	if err != nil {
		inst.logErr(err)
		return
	}
	streamPort, err := inst.sendTriggerMsg(wsMsg)
	if err != nil {
		inst.logErr(err)
		return
	}
	inst.streamTracker.register(msg, streamPort)
	defer inst.streamTracker.remove(msg.ID)
	if err := inst.runStream(msg.ID, streamPort); err != nil {
		inst.logErr(err)
	}
}

// sendTriggerMsg sends the trigger message over the main WS connection and
// returns the stream port assigned by the server.
func (inst *WidgetService) sendTriggerMsg(wsMsg *models.WSMessage) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := inst.setup.GetWSClient()
	if err := client.Connect(ctx); err != nil {
		return 0, err
	}
	defer client.Close()
	if err := client.SendJSON(ctx, wsMsg); err != nil {
		return 0, err
	}
	return inst.readStreamPort(ctx, client)
}

// readStreamPort reads the server's response and extracts the assigned port.
func (inst *WidgetService) readStreamPort(ctx context.Context, client *wsclient.Client) (int, error) {
	var resp models.WSMessage
	if err := client.ReadJSON(ctx, &resp); err != nil {
		return 0, err
	}
	if resp.Error != "" {
		return 0, fmt.Errorf("trigger rejected: %s", resp.Error)
	}
	var portData map[string]int
	if err := json.Unmarshal(resp.Data, &portData); err != nil {
		return 0, fmt.Errorf("malformed port response: %w", err)
	}
	if portData["port"] == 0 {
		return 0, fmt.Errorf("no port in trigger response")
	}
	return portData["port"], nil
}

// runStream connects to the assigned stream socket and drains it.
func (inst *WidgetService) runStream(msgID string, streamPort int) error {
	endpoint, err := inst.setup.GetEndpointResolver().Resolve()
	if err != nil {
		return err
	}
	streamURL := fmt.Sprintf("ws://%s:%d/ws", endpoint.Hostname, streamPort)
	client := wsclient.NewWSClient(streamURL, inst.setup.GetToken(), inst.logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		return err
	}
	defer client.Close()
	return inst.drainStream(ctx, client, msgID)
}

// drainStream sends the opening ACK, prints messages until STREAM-END, then
// sends the closing ACK.
func (inst *WidgetService) drainStream(ctx context.Context, client *wsclient.Client, msgID string) error {
	if err := client.SendText(ctx, TRIGGER_PROTOCOL_ACKNOWLEDGEMENT); err != nil {
		return err
	}
	inst.streamTracker.updateStatus(msgID, models.TriggerOperationStatusAck)
	for {
		text, err := client.ReadText(ctx)
		if err != nil {
			return err
		}
		if text == TRIGGER_STREAM_END {
			break
		}
		if inst.logger != nil {
			inst.logger.Info(fmt.Sprintf("[trigger %s] %s", msgID, text))
		}
	}
	_ = client.SendText(ctx, TRIGGER_PROTOCOL_ACKNOWLEDGEMENT)
	inst.streamTracker.updateStatus(msgID, models.TriggerOperationStatusDone)
	return nil
}

func (inst *WidgetService) SetOnClosed(close func()) {
	if inst.uiHolder != nil {
		inst.uiHolder.SetOnClosed(close)
	}

	if inst.dashboardHolder != nil {
		inst.dashboardHolder.SetOnClosed(close)
	}
}

func (inst *WidgetService) CreateProjectUI(vp *viewproject.ViewProject, props *properties.Properties) error {
	if vp == nil || len(vp.View.Diagrams) == 0 || len(vp.View.Diagrams[0].Setup) == 0 {
		return fmt.Errorf("invalid view project: missing diagrams/setup")
	}
	inst.uiHolder = ui.NewProjectUIHolder(vp, inst.renderer, inst.drawerFactory, props, inst.variableContainer)
	if inst.triggerSenderFactory != nil {
		inst.uiHolder.SetTriggerSenderFactory(inst.triggerSenderFactory)
	}
	err := inst.uiHolder.Create()
	if err != nil {
		return err
	}
	return nil
}

func (inst *WidgetService) CreateDashboardProjectUI(dp *viewproject.DashboardProject, props *properties.Properties) error {
	if dp == nil || dp.IsEmpty() {
		return fmt.Errorf("invalid dashboard project: the project is empty")
	}
	inst.dashboardHolder = ui.NewDashboardUIHolder(dp, inst.renderer, inst.drawerFactory, props, inst.jobRouter)
	if inst.triggerSenderFactory != nil {
		inst.dashboardHolder.SetTriggerSenderFactory(inst.triggerSenderFactory)
	}
	err := inst.dashboardHolder.Create(dp)
	if err != nil {
		return err
	}
	return nil
}

func (inst *WidgetService) Update(vp *viewproject.ViewProject) error {
	return nil
}

func (inst *WidgetService) Run() {
	if inst.uiHolder != nil {
		inst.uiHolder.Run()
	}

	if inst.dashboardHolder != nil {
		inst.dashboardHolder.Run()
	}
}

func (inst *WidgetService) DispatchValue(jobID string, varName string, t time.Time, value any) {
	if inst.jobRouter != nil {
		inst.jobRouter.Push(jobID, varName, t, value)
	}
	inst.variableContainer.Push(varName, t, value)
}
