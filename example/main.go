package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yadukrishnan2004/antrelay-sdk/client"
)

func processOrder(_ context.Context, input []byte) ([]byte, error) {
	log.Printf("processing order: %s", input)
	return []byte(`{"status":"charged"}`), nil
}

func sendEmail(_ context.Context, input []byte) ([]byte, error) {
	log.Printf("sending email: %s", input)
	return []byte(`{"status":"sent"}`), nil
}

func main(){
c := client.New(
    client.Config{
        ServerURL: "http://localhost:8080",
        Queue:     "orders-queue",
    },
    client.WithPollInterval(500*time.Millisecond),
    client.WithHandlerTimeout(30*time.Second),
    client.WithMaxPollRetries(5),
)

	if err := c.Register("processOrder", processOrder); err != nil {
		fmt.Fprintf(os.Stderr, "failed to register handler: %v\n", err)
		os.Exit(1)
	}

	if err := c.Register("sendEmail", sendEmail); err != nil {
		fmt.Fprintf(os.Stderr, "failed to register handler: %v\n", err)
		os.Exit(1)
	}

	c.Start()
	log.Println("worker started — waiting for tasks")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	c.Stop()
}