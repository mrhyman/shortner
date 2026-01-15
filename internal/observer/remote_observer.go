// internal/observer/remote_observer.go
package observer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mrhyman/shortner/internal/logger"
)

type RemoteObserver struct {
	url    string
	client *http.Client
}

func NewRemoteObserver(url string) *RemoteObserver {
	return &RemoteObserver{
		url: url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (o *RemoteObserver) OnRequest(event Event) {
	data, _ := json.Marshal(event)

	resp, err := o.client.Post(o.url, "application/json", bytes.NewReader(data))
	if err != nil {
		logger.FromContext(resp.Request.Context()).With("err", err.Error())
		return
	}
	resp.Body.Close()
}
