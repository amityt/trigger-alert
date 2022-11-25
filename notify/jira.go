package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"trigger-alert/types"
)

type JiraNotify struct {
	Fields Field `json:"fields"`
}

type JiraResponse struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

type Field struct {
	Project struct {
		Key string `json:"key"`
	} `json:"project"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	IssueType   struct {
		Name string `json:"name"`
	} `json:"issuetype"`
}

// NotifySlack notifies slack about the summary of DB upgrade
func NotifyJira(jiraUrl string, expDetails types.ExperimentDetails) (string, error) {
	var jiraResponse JiraResponse
	notificationDetails := Field{
		Project: struct {
			Key string `json:"key"`
		}{Key: "HG"},
		Summary:     "Chaos Experiment Failed: " + expDetails.ExperimentName,
		Description: "Experiment: " + expDetails.ExperimentName + " FailStep: " + expDetails.FailStep,
		IssueType: struct {
			Name string `json:"name"`
		}{Name: "Bug"},
	}

	notifyMsg := &JiraNotify{
		Fields: notificationDetails,
	}
	data, err := json.Marshal(notifyMsg)
	fmt.Println(string(data))
	if err != nil {
		return "", err
	}

	request, err := http.NewRequest(http.MethodPost, jiraUrl, bytes.NewBuffer(data))

	if err != nil {
		return "", err
	}

	request.Header = http.Header{
		"Content-Type": []string{"application/json"},
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(bodyBytes, &jiraResponse)

	fmt.Println(jiraResponse.Key)

	return jiraResponse.Key, nil
}
