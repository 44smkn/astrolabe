package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type Trigger interface {
	Pull() error
	IsCompleted() (bool, error)
}

type WebhookOnSpinnaker struct {
	URL           string            `yaml:"url"`
	Body          map[string]string `yaml:"body"`
	SpinCliConfig string            `yaml:"spinCliConfig"`
	EventId       string
	GateEndpoint  string
	SpinToken     *oauth2.Token
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
			CachedToken *oauth2.Token `yaml:"cachedToken"`
		} `yaml:"oauth2"`
	} `yaml:"auth"`
}

func NewTrigger(trigger PipelineTrigger) (Trigger, error) {
	switch trigger.Enabled {
	case "webhook":
		var config SpinCliConfig
		if err := DeserializeYamlFile(trigger.Webhook.SpinCliConfig, config); err != nil {
			return nil, err
		}
		webhook := trigger.Webhook
		webhook.GateEndpoint = config.Gate.Endpoint
		webhook.SpinToken = config.Auth.Oauth2.CachedToken
		return &webhook, nil
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

// https://spinnaker.io/guides/user/pipeline/searching/
func (w *WebhookOnSpinnaker) IsCompleted() (bool, error) {
	req, err := http.NewRequest(http.MethodGet, w.GateEndpoint, nil)
	if err != nil {
		return false, err
	}
	w.SpinToken.SetAuthHeader(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	// TODO: responseからステータスを読み取る
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
