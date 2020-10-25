package workflow

import (
	"time"

	"github.com/calebamiles/example-money-making-fortune-api/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	// TaskList is the queue to use for GetFortune execution
	TaskList = "getFigletedFortuneTaskList"
)

// GetFigletizedFortune returns a new fortune, or the default fortune
func GetFigletizedFortune(ctx workflow.Context) (string, error) {
	ao := workflow.ActivityOptions{
		TaskList:               TaskList,
		ScheduleToCloseTimeout: time.Second * 60,
		ScheduleToStartTimeout: time.Second * 60,
		StartToCloseTimeout:    time.Second * 60,
		HeartbeatTimeout:       time.Second * 10,
		WaitForCancellation:    false,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	future := workflow.ExecuteActivity(ctx, activity.GetFigletizedFortune)
	var fortune string
	if err := future.Get(ctx, &fortune); err != nil {

		workflow.GetLogger(ctx).Error("Executing GetFigletizedFortune activity", zap.Error(err))
		return "", err
	}

	workflow.GetLogger(ctx).Info("GetFigletizedFortune workflow done")
	return fortune, nil
}
