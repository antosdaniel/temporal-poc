package workflows

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func ProcessPayroll(ctx workflow.Context, payrollID string) error {
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var canBeProcessed bool
	err := workflow.ExecuteActivity(ctx, CanPayrollBeProcessed, payrollID).Get(ctx, &canBeProcessed)
	if err != nil {
		return err
	}
	if !canBeProcessed {
		return nil
	}

	// While we process FPS, we start processing payments.
	processPayments := workflow.ExecuteChildWorkflow(ctx, ProcessPayments, payrollID)

	// Report FPS.
	var fpsReference FPSReportReference
	err = workflow.ExecuteActivity(ctx, ReportFPS, payrollID).Get(ctx, &fpsReference)
	if err != nil {
		return err
	}

	// HMRC can take its sweet time to validate FPS. We await until it tells us if FPS was successful or not.
	// In realistic scenario, we would probably start this in a separate workflow, so it doesn't block other actions.
	checkStatusCtx := workflow.WithRetryPolicy(ctx, temporal.RetryPolicy{
		MaximumInterval: time.Second,
		// In reality, it would look more like this:
		//InitialInterval:    time.Minute,
		//BackoffCoefficient: 5,
		//MaximumInterval:    time.Hour * 24,
	})
	for {
		var fpsStatus FPSReportStatus
		err = workflow.ExecuteActivity(checkStatusCtx, CheckFPSReport, payrollID).Get(checkStatusCtx, &fpsStatus)
		if err != nil {
			return err
		}
		if fpsStatus.StillPending {
			continue
		}
		if !fpsStatus.WasSuccessFull {
			return errors.New("FPS has business errors")
		}

		err = workflow.ExecuteActivity(checkStatusCtx, MarkFPSAsSuccessful, payrollID).Get(checkStatusCtx, nil)
		if err != nil {
			return err
		}
		break
	}

	// We are pretending that after successful FPS submission, we send payslips to employees.
	// Again, it probably would be its own workflow.
	err = workflow.ExecuteActivity(checkStatusCtx, SendDocuments, payrollID).Get(checkStatusCtx, nil)
	if err != nil {
		return err
	}

	return workflow.Await(ctx, func() bool {
		if !processPayments.IsReady() {
			return false
		}
		return true
	})
}

func CanPayrollBeProcessed(_ context.Context, payrollID string) (bool, error) {
	return true, nil
}

type FPSReportReference string

func ReportFPS(_ context.Context, payrollID string) (FPSReportReference, error) {
	time.Sleep(time.Second * 9)

	if err := failXOutOf10Times(5); err != nil {
		return "", err
	}

	return FPSReportReference("fps-" + payrollID), nil
}

type FPSReportStatus struct {
	StillPending   bool
	WasSuccessFull bool
	Details        string
}

func CheckFPSReport(_ context.Context, reference FPSReportReference) (FPSReportStatus, error) {
	time.Sleep(time.Second)

	if err := failXOutOf10Times(9); err != nil {
		return FPSReportStatus{StillPending: true}, err
	}

	if err := failXOutOf10Times(5); err != nil {
		return FPSReportStatus{WasSuccessFull: false, Details: "HMRC is down"}, err
	}

	return FPSReportStatus{WasSuccessFull: true}, nil
}

func MarkFPSAsSuccessful(_ context.Context, payrollID string) error {
	fmt.Printf("FPS for payroll %q was accepted by HMRC\n", payrollID)
	return nil
}

func SendDocuments(_ context.Context, payrollID string) error {
	fmt.Println("Sending payslips to employees...")
	return nil
}
