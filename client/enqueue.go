package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type EnqueueResult struct{
	TaskID string
	Queue string
	Function string
}

type EnqueueRequest struct{
	FunctionName string  `json:"function_name"`
	Input json.RawMessage `json:"input"`
	MaxRetries int `json:"max_retrys"`
}


func (c *Client) Enqueue(functionName string, input []byte)(*EnqueueResult, error){
	return c.EnqueueToQueue(functionName,input,c.config.Queue)
}

func (c *Client) EnqueueToQueue(functionName string, input []byte, queue string)(*EnqueueResult, error){
	if functionName == "" {
		return nil, &ValidationError{
			Field:   "functionName",
			Message: "cannot be empty",
		}
	}

	if len(input) == 0 {
		input = []byte(`{}`) 
	}

	req := EnqueueRequest{
		FunctionName: functionName,
		Input:        input,
		MaxRetries:   c.opts.maxPollRetries,
	}

	body, err := json.Marshal(req) 
	if err != nil {
		return nil, fmt.Errorf("clinet : failed to encode request: %w", err)
	}

	url := fmt.Sprintf("%s/queues/%s/tasks", c.config.ServerURL, queue)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("client: failed to enqueue task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("client: server returned unexpected status: %d", resp.StatusCode)
	}

	var taskResp struct {
		ID           string `json:"ID"`
		FunctionName string `json:"FunctionName"`
		Queue        string `json:"Queue"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return nil, fmt.Errorf("client: failed to decode response: %w", err)
	}

	return &EnqueueResult{
		TaskID:   taskResp.ID,
		Queue:    taskResp.Queue,
		Function: taskResp.FunctionName,
	}, nil
}
