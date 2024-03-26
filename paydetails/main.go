package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"temporal-poc/paydetails/workflows"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
)

const taskQueue = "backgroundcheck-boilerplate-task-queue-local"

func main() {
	fmt.Println("Starting worker...")
	ctx := context.Background()

	clientOptions := client.Options{
		Namespace: "default",
	}
	temporalClient, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create a Temporal Client", err)
	}
	defer temporalClient.Close()

	err = registerSchedules(ctx, temporalClient.ScheduleClient())
	if err != nil && !errors.Is(err, temporal.ErrScheduleAlreadyRunning) {
		log.Fatalln("Unable to register schedules", err)
	}

	w := worker.New(temporalClient, taskQueue, worker.Options{})
	registerWorkflows(w)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start the Worker Process", err)
	}
}

func registerWorkflows(w worker.Worker) {
	// Workflows
	w.RegisterWorkflow(workflows.BackgroundCheck)
	w.RegisterWorkflow(workflows.SyncDataFromBob)

	// Activities
	w.RegisterActivity(workflows.SSNTraceActivity)
	w.RegisterActivity(workflows.PullData)
	w.RegisterActivity(workflows.StoreData)
}

func registerSchedules(ctx context.Context, c client.ScheduleClient) error {
	_, err := c.Create(ctx, client.ScheduleOptions{
		ID: "sync-data-from-bob-every-minute",
		Spec: client.ScheduleSpec{
			CronExpressions: []string{"* * * * *"},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        "sync-data-from-bob-every-minute",
			Workflow:  workflows.SyncDataFromBob,
			TaskQueue: taskQueue,
		},
		Overlap: enums.SCHEDULE_OVERLAP_POLICY_SKIP,
	})
	if err != nil && !alreadyScheduled(err) {
		return err
	}

	// More schedules...

	return nil
}

func alreadyScheduled(err error) bool {
	return errors.Is(err, temporal.ErrScheduleAlreadyRunning)
}
