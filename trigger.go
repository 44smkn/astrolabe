package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

type Trigger interface {
	Pull() error
}

type Webhook struct {
	URL  string            `yaml:"url"`
	Body map[string]string `yaml:"body"`
}

type SpinCli struct {
	CertFilePath string `yaml:"certFilePath"`
}

func NewTrigger(trigger PipelineTrigger) (Trigger, error) {
	switch trigger.Enabled {
	case "webhook":
		return &Webhook{
			trigger.Webhook.URL,
			trigger.Webhook.Body,
		}, nil
	case "spinCli":
		return &SpinCli{trigger.SpinCli.CertFilePath}, nil
	default:
		return nil, errors.New("The specified trigger does not exist")
	}
}

func (w *Webhook) Pull() error {
	buf, err := json.Marshal(w.Body)
	if err != nil {
		return errors.Wrap(err, "failed to serialize body of the webhook request.")
	}
	_, err = http.Post(w.URL, "application/json", bytes.NewReader(buf))
	if err != nil {
		return errors.Wrap(err, "webhook request is failed.")
	}
	return nil
}

func (s *SpinCli) Pull() error {
	//TODO: implement
	return nil
}
