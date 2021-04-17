package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type SlackRequestBody struct {
	Attachments []SlackAttachment `json:"attachments"`
}

type SlackAttachment struct {
	Title     string `json:"title"`
	TitleLink string `json:"title_link"`
}

func SendSlackNotification(slackLink, title, link string) error {
	slkMsg := SlackRequestBody{
		Attachments: []SlackAttachment{
			SlackAttachment{
				Title:     title,
				TitleLink: link,
			},
		},
	}
	slackBody, _ := json.Marshal(slkMsg)
	req, err := http.NewRequest(http.MethodPost, slackLink, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		return errors.New("Non-ok response returned from Slack")
	}
	return nil
}
