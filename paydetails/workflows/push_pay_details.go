package workflows

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type PushPayDetailsInput struct {
	CompanyID string
	PayslipID string
}

func PushPayDetails(ctx workflow.Context, input PushPayDetailsInput) error {
	// Most actions should always complete. We're creating infinite-retry here.
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})

	// I'm not sure if this shouldn't be an action as well?
	// If we wanted data to be re-fetched on retry, it would have to be under `PushPayDetailsToBob`.
	// Since this is passed to each activity below, pay details will be recorded within temporal. This has its issues:
	// - PII could end up in temporal
	// - If input (payDetails) is big enough, this could affect performance of temporal.
	payDetails, err := pullData(input.CompanyID, input.PayslipID)
	if err != nil {
		return err
	}

	// User would like to know that we're trying to send something.
	err = workflow.ExecuteActivity(ctx, MarkPayDetailsAsBeingSent, payDetails).Get(ctx, nil)
	if err != nil {
		return err
	}

	// If we can't push the data to Bob, we'll retry up to 5 times.
	limitedRetryCtx := workflow.WithRetryPolicy(ctx, temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Second * 100,
		MaximumAttempts:    5,
	})
	err = workflow.ExecuteActivity(limitedRetryCtx, PushPayDetailsToBob, payDetails).Get(limitedRetryCtx, nil)
	if err != nil {
		// No dice. We'll mark entire workflow as failed.
		_ = workflow.ExecuteActivity(ctx, MarkPayDetailsAsFailed, payDetails).Get(ctx, nil)
		return err
	}

	// Success!
	_ = workflow.ExecuteActivity(ctx, MarkPayDetailsAsSent, payDetails).Get(ctx, nil)

	return nil
}

type PayDetails struct {
	CompanyID string
	PayslipID string
	Payday    time.Time
	FirstName string
	LastName  string
	Salary    int
}

func pullData(companyID, payslipID string) (PayDetails, error) {
	fmt.Println("fetching data...")
	return PayDetails{
		CompanyID: companyID,
		PayslipID: payslipID,
		Payday:    time.Time{},
		FirstName: "Joe",
		LastName:  "Smith",
		Salary:    10_000_00,
	}, nil
}

func PushPayDetailsToBob(ctx context.Context, payDetails PayDetails) error {
	time.Sleep(time.Second)
	if rand.IntN(3) == 0 { // 33% of the times succeeds
		return nil
	}
	return errors.New("failed to push pay details to Bob")
}

func MarkPayDetailsAsBeingSent(ctx context.Context, payDetails PayDetails) error {
	fmt.Println("Trying to send pay details")
	return nil
}

func MarkPayDetailsAsFailed(ctx context.Context, payDetails PayDetails) error {
	fmt.Println("Pay details failed")
	return nil
}

func MarkPayDetailsAsSent(ctx context.Context, payDetails PayDetails) error {
	fmt.Println("Pay details sent")
	return nil
}
