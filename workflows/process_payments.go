package workflows

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
)

func ProcessPayments(ctx workflow.Context, payrollID string) error {
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Find payments we should execute.
	var payments Payments
	err := workflow.ExecuteActivity(ctx, FindPayments, payrollID).Get(ctx, &payments)
	if err != nil {
		return err
	}

	processedPayments := 0
	for _, payment := range payments {
		workflow.GoNamed(ctx, payment.PaymentID, func(ctx workflow.Context) {
			// Each payment has to be successfully scheduled.
			// Even if there are not enough funds, we will retry until it succeeds.
			err := workflow.ExecuteActivity(ctx, SchedulePayment, payment).Get(ctx, nil)
			if err != nil {
				return
			}

			// We care about them being actually paid. We wait for that.
			var isPaid bool
			for !isPaid {
				err = workflow.ExecuteActivity(ctx, IsPaymentPaid, payment.PaymentID).Get(ctx, &isPaid)
				if err != nil {
					return
				}
			}

			// Once payment was paid, we mark it as such and reconcile it in accounting integration.
			err = workflow.ExecuteActivity(ctx, ReconcileInAccountingIntegration, payment.PaymentID).Get(ctx, nil)
			if err != nil {
				return
			}

			processedPayments++
		})
	}

	return workflow.Await(ctx, func() bool {
		return processedPayments == len(payments)
	})
}

type (
	Payments []Payment
	Payment  struct {
		PaymentID string
		Amount    int
	}
)

func FindPayments(ctx context.Context, payrollID string) (Payments, error) {
	return Payments{
		{PaymentID: "1", Amount: 100},
		{PaymentID: "2", Amount: 200},
		{PaymentID: "3", Amount: 300},
	}, nil
}

func SchedulePayment(ctx context.Context, payment Payment) error {
	time.Sleep(time.Second)
	return failXOutOf10Times(3)
}

func IsPaymentPaid(ctx context.Context, paymentID string) (bool, error) {
	time.Sleep(time.Second)
	return true, failXOutOf10Times(3)
}

func ReconcileInAccountingIntegration(ctx context.Context, paymentID string) error {
	time.Sleep(time.Second)
	return nil
}
