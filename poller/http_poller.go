package poller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yadukrishnan2004/antrelay-sdk/task"
)

type HTTPPoller struct{
	serverURL string
	httpClient *http.Client
}

func NewHTTP(serverURL string)*HTTPPoller{
	return &HTTPPoller{
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: 10*time.Second,
		},
	}
}

func (p *HTTPPoller) Poll(queue string)(*task.Task,error){

	url:=fmt.Sprintf("%s/queues/%s/tasks/next",p.serverURL,queue)

	resp,err:=p.httpClient.Get(url)
	if err != nil{
		return nil,fmt.Errorf("poller: http request faild: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent{
		return nil,nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil,fmt.Errorf("poller : unexpected status code: %d", resp.StatusCode)
	}

	var t task.Task
	if err:= json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil,fmt.Errorf("poller : failed to decode task: %w", err)
	}

	return &t, nil

}


func (p *HTTPPoller) Requeue(t *task.Task) {
	url := fmt.Sprintf("%s/tasks/%s/requeue", p.serverURL,t.ID)
	resp,err := p.httpClient.Post(url,"application/json",nil)
	if err != nil {
		fmt.Printf("poller: requeue request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
}

