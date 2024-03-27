package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	workflows2 "temporal-poc/workflows"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
)

const taskQueue = "default"

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

	go pushPayDayDetails(ctx, temporalClient)
	go pushPayDayDetails(ctx, temporalClient)
	go processPayroll(ctx, temporalClient)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start the Worker Process", err)
	}
}

func registerWorkflows(w worker.Worker) {
	// Very simple example. Probably not the best case though. But I wanted to show cron scheduling.
	w.RegisterWorkflow(workflows2.SyncDataFromBob)
	w.RegisterActivity(workflows2.PullData)
	w.RegisterActivity(workflows2.StoreData)

	// Pushing pay details in a *similar* way we did doc-sender. A lot less plumbing!
	w.RegisterWorkflow(workflows2.PushPayDetails)
	w.RegisterActivity(workflows2.PushPayDetailsToBob)
	w.RegisterActivity(workflows2.MarkPayDetailsAsBeingSent)
	w.RegisterActivity(workflows2.MarkPayDetailsAsFailed)
	w.RegisterActivity(workflows2.MarkPayDetailsAsSent)

	// A lot more complicated process.
	w.RegisterWorkflow(workflows2.ProcessPayroll)
	w.RegisterActivity(workflows2.CanPayrollBeProcessed)
	w.RegisterActivity(workflows2.ReportFPS)
	w.RegisterActivity(workflows2.CheckFPSReport)
	w.RegisterActivity(workflows2.MarkFPSAsSuccessful)
	w.RegisterActivity(workflows2.SendDocuments)
}

func registerSchedules(ctx context.Context, c client.ScheduleClient) error {
	_, err := c.Create(ctx, client.ScheduleOptions{
		ID: "sync-data-from-bob-every-minute",
		Spec: client.ScheduleSpec{
			CronExpressions: []string{"* * * * *"},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        "sync-data-from-bob-every-minute",
			Workflow:  workflows2.SyncDataFromBob,
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

func pushPayDayDetails(ctx context.Context, c client.Client,) {
	input := workflows2.PushPayDetailsInput{
		CompanyID: "company-id",
		PayslipID: "payslip-id",
	}
	workflowOptions := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("push-pay-details-%s-%s", input.CompanyID, input.PayslipID),
		TaskQueue:             taskQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	_, err := c.ExecuteWorkflow(ctx, workflowOptions, workflows2.PushPayDetails, input)
	if err != nil {
		log.Fatalln("Unable to push pay details", err)
	}
}

func processPayroll(ctx context.Context, c client.Client,) {
	payrollID := "payroll-id"
	workflowOptions := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("process-payroll-%s", payrollID),
		TaskQueue:             taskQueue,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
	}
	_, err := c.ExecuteWorkflow(ctx, workflowOptions, workflows2.ProcessPayroll, payrollID)
	if err != nil {
		log.Fatalln("Unable to push pay details", err)
	}
}
