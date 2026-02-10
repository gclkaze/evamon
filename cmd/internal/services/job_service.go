package services

import (
	"context"
	"encoding/json"

	"github.com/gclkaze/evamon/cmd/internal/models"
	"github.com/gclkaze/evamon/cmd/internal/output"
)

type JobService struct {
	logger output.Printer
	setup  MainSetup
}

func NewJobService() *JobService {
	return &JobService{}
}

func (inst *JobService) SetSetup(setup MainSetup) {
	inst.setup = setup
	inst.logger = setup.GetPrinter()
}

func (inst *JobService) AddJob(job *models.JobAddRequest, widgetPath string) error {
	client := inst.setup.GetWSClient()
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		return err
	}
	defer client.Close()
	raw, err := json.Marshal(job)
	if err != nil {
		return err
	}
	msg := models.WSMessage{
		Type: "job.add",
		Data: raw,
	}

	if err := client.SendJSON(ctx, msg); err != nil {
		return err
	}

	var resp models.WSMessage
	if err := client.ReadJSON(ctx, &resp); err != nil {
		return err
	}
	return nil
}
