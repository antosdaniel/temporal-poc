package workflows

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

// SyncDataFromBob is just an example. It probably is not a good idea to break sync into two activities, as all
// of this input/output data would be stored in Temporal's database.
func SyncDataFromBob(ctx workflow.Context) error {
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var data DataFromBob
	err := workflow.ExecuteActivity(ctx, PullData, nil).Get(ctx, &data)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, StoreData, data).Get(ctx, &data)
	if err != nil {
		return err
	}

	return nil
}

type DataFromBob struct {
	EmployeeID string
	Salary     int
}

func PullData(_ context.Context) (DataFromBob, error) {
	return DataFromBob{
		EmployeeID: "employee-1",
		Salary:     10_000_00,
	}, nil
}

func StoreData(_ context.Context, data DataFromBob) error {
	fmt.Printf("Storing data: %v\n", data)
	return nil
}
