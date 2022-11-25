package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
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

func AttachLogJira(jiraUrl string, issueID string) (string, error) {

	b, w := createMultipartFormData("file", issueID+".txt")
	fmt.Println(jiraUrl)
	request, err := http.NewRequest(http.MethodPost, jiraUrl, &b)

	if err != nil {
		return "", err
	}
	request.Header.Add("Content-Type", w.FormDataContentType())
	request.Header.Add("X-Atlassian-Token", "no-check")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	sb := string(bodyBytes)
	log.Printf(sb)

	e := os.RemoveAll(issueID + ".txt")
	if e != nil {
		log.Fatal(e)
	}
	return "", nil
}

func createMultipartFormData(fieldName, fileName string) (bytes.Buffer, *multipart.Writer) {
	b := new(bytes.Buffer)
	var err error
	w := multipart.NewWriter(b)
	var fw io.Writer
	file := mustOpen(fileName)
	if fw, err = w.CreateFormFile(fieldName, file.Name()); err != nil {
		fmt.Println("Error creating writer: %v", err)
	}
	if _, err = io.Copy(fw, file); err != nil {
		fmt.Println("Error with io.Copy: %v", err)
	}
	w.Close()
	return *b, w
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		fmt.Println(err)
	}
	return r
}
