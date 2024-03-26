package workflows

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
)

// BackgroundCheck is your custom Workflow Definition.
func BackgroundCheck(ctx workflow.Context, param string) (string, error) {
	// Define the Activity Execution options
	// StartToCloseTimeout or ScheduleToCloseTimeout must be set
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	// Execute the Activity synchronously (wait for the result before proceeding)
	var ssnTraceResult string
	err := workflow.ExecuteActivity(ctx, SSNTraceActivity, param).Get(ctx, &ssnTraceResult)
	if err != nil {
		return "", err
	}
	// Make the results of the Workflow available
	return ssnTraceResult, nil
}

func SSNTraceActivity(ctx context.Context, param string) (*string, error) {
	// This is where a call to another service is made
	// Here we are pretending that the service that does SSNTrace returned "pass"
	result := "pass"
	return &result, nil
}
