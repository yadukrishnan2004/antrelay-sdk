package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yadukrishnan2004/antrelay-sdk/task"
)


type HTTPReporter struct{
	serverURL string
	httpClient *http.Client
}


func NewHTTP(serverURL string) *HTTPReporter{
	return &HTTPReporter{
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (r *HTTPReporter) Report(result *task.Result) error {
	url := fmt.Sprintf("%s/tasks/%s/result", r.serverURL, result.TaskID)

	body, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("reporter : failed to encode result: %w", err)
	}

	resp, err := r.httpClient.Post(url, "application/json",bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("reporter: http request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("reporter: unexpected status code: %d", resp.StatusCode)
	}

	return nil
}