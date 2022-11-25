package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"trigger-alert/types"
)

type SlackNotify struct {
	Blocks []Section `json:"blocks"`
}

type Section struct {
	Type string      `json:"type"`
	Text TextDetails `json:"text,omitempty"`
}

type TextDetails struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NotifySlack notifies slack about the summary of DB upgrade
func NotifySlack(slackURL string, jiraTicketLink string, expDetails types.ExperimentDetails) error {

	var notificationDetails []Section

	expName := Section{
		Type: "section",
		Text: TextDetails{
			Type: "mrkdwn",
			Text: "Experiment Name:* " + expDetails.ExperimentName,
		},
	}
	notificationDetails = append(notificationDetails, expName)

	engineName := Section{
		Type: "section",
		Text: TextDetails{
			Type: "mrkdwn",
			Text: "*Engine Name:* " + expDetails.EngineName,
		},
	}
	notificationDetails = append(notificationDetails, engineName)

	failStep := Section{
		Type: "section",
		Text: TextDetails{
			Type: "mrkdwn",
			Text: "*FailStep:* " + expDetails.FailStep,
		},
	}
	notificationDetails = append(notificationDetails, failStep)

	phase := Section{
		Type: "section",
		Text: TextDetails{
			Type: "mrkdwn",
			Text: "*Phase:* " + expDetails.Phase,
		},
	}
	notificationDetails = append(notificationDetails, phase)

	probeSuccessPercentage := Section{
		Type: "section",
		Text: TextDetails{
			Type: "mrkdwn",
			Text: "*ProbeSuccessPercentage:* " + expDetails.ProbeSuccessPercentage,
		},
	}
	notificationDetails = append(notificationDetails, probeSuccessPercentage)

	linkJira := Section{
		Type: "section",
		Text: TextDetails{
			Type: "mrkdwn",
			Text: "*Jira Ticket Link:* " + jiraTicketLink,
		},
	}
	notificationDetails = append(notificationDetails, linkJira)

	notifyMsg := &SlackNotify{
		Blocks: notificationDetails,
	}
	data, err := json.Marshal(notifyMsg)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, slackURL, bytes.NewBuffer(data))

	if err != nil {
		return err
	}

	request.Header = http.Header{
		"Content-Type": []string{"application/json"},
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("error in notifying slack")
	}

	return nil
}
