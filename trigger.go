package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type Trigger interface {
	Pull() error
	IsCompleted() (bool, error)
}

type WebhookOnSpinnaker struct {
	URL          string            `yaml:"url"`
	Body         map[string]string `yaml:"body"`
	CertFilePath string            `yaml:"certFilePath"`
	EventId      string
}

type SpinCli struct {
	CertFilePath string `yaml:"certFilePath"`
}

type SpinCliConfig struct {
	Gate struct {
		Endpoint string `yaml:"endpoint"`
	} `yaml:"gate"`
	Auth struct {
		Oauth2 struct {
			CachedToken struct {
				AccessToken  string `yaml:"access_token"`
				TokenType    string `yaml:"token_type"`
				RefreshToken string `yaml:"refresh_token"`
				Expiry       struct {
				} `yaml:"expiry"`
			} `yaml:"cachedToken"`
		} `yaml:"oauth2"`
	} `yaml:"auth"`
}

func NewTrigger(trigger PipelineTrigger) (Trigger, error) {
	switch trigger.Enabled {
	case "webhook":
		return &WebhookOnSpinnaker{
			URL:          trigger.Webhook.URL,
			Body:         trigger.Webhook.Body,
			CertFilePath: trigger.Webhook.CertFilePath,
			EventId:      "",
		}, nil
	case "spinCli":
		return &SpinCli{trigger.SpinCli.CertFilePath}, nil
	default:
		return nil, errors.New("The specified trigger does not exist")
	}
}

func (w *WebhookOnSpinnaker) Pull() error {
	buf, err := json.Marshal(w.Body)
	if err != nil {
		return errors.Wrap(err, "failed to serialize body of the webhook request.")
	}
	resp, err := http.Post(w.URL, "application/json", bytes.NewReader(buf))
	if err != nil {
		return errors.Wrap(err, "webhook request is failed.")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	m := make(map[string]string)
	err = json.Unmarshal(body, &m)
	if err != nil {
		return err
	}
	w.EventId = m["eventId"]

	return nil
}

func (w *WebhookOnSpinnaker) IsCompleted() (bool, error) {

	return true, nil
}

func (s *SpinCli) Pull() error {
	//TODO: implement
	return nil
}

func (s *SpinCli) IsCompleted() (bool, error) {
	//TODO: implement
	return false, nil
}
