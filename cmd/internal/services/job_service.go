package services

import (
	"context"

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
}

func (inst *JobService) AddJob(job *models.JobAddRequest, widgetPath string) error {
	client := inst.setup.GetWSClient()
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		return err
	}
	defer client.Close()

	msg := models.WSMessage{
		Type: "job.add",
		Data: job,
	}

	// 7) Send request
	if err := client.SendJSON(ctx, msg); err != nil {
		return err
	}

	// 8) Read response
	var resp models.WSMessage
	if err := client.ReadJSON(ctx, &resp); err != nil {
		return err
	}
	return nil
}
