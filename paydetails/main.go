package main

import (
	"fmt"
	"log"

	"temporal-poc/paydetails/workflows"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	fmt.Println("Starting worker...")
	// Initialize a Temporal Client
	// Specify the Namespace in the Client options
	clientOptions := client.Options{
		Namespace: "default",
	}
	temporalClient, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create a Temporal Client", err)
	}
	defer temporalClient.Close()
	// Create a new Worker
	yourWorker := worker.New(temporalClient, "backgroundcheck-boilerplate-task-queue-local", worker.Options{})
	// Register Workflows
	yourWorker.RegisterWorkflow(workflows.BackgroundCheck)
	// Register Activities
	yourWorker.RegisterActivity(workflows.SSNTraceActivity)
	// Start the Worker Process
	err = yourWorker.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start the Worker Process", err)
	}
}
