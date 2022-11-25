package k8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"
	litmusV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/typed/litmuschaos/v1alpha1"
	v1alpha12 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/typed/litmuschaos/v1alpha1"
	"github.com/sirupsen/logrus"
	"io"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"strings"
	"trigger-alert/notify"
	"trigger-alert/types"
)

var KubeConfig *string

type TriggerResponseConfig struct {
	SlackURL        string
	JiraURL         string
	ExperimentRunID string
	Namespace       string
}

func LoadEnv() TriggerResponseConfig {
	return TriggerResponseConfig{
		SlackURL:        os.Getenv("SLACK_URL"),
		JiraURL:         os.Getenv("JIRA_URL"),
		ExperimentRunID: os.Getenv("EXPERIMENT_RUN_ID"),
		Namespace:       os.Getenv("NAMESPACE"),
	}
}

func (c *TriggerResponseConfig) GetChaosEngines(namespace string, workflowRunID string) error {
	ctx := context.TODO()

	//Define the GVR
	resourceType := schema.GroupVersionResource{
		Group:    "litmuschaos.io",
		Version:  "v1alpha1",
		Resource: "chaosengines",
	}

	//Generate the dynamic client
	dynamicClient, err := GetDynamicClient()
	if err != nil {
		return errors.New("failed to get dynamic client, error: " + err.Error())
	}

	listOption := v1.ListOptions{}

	listOption.LabelSelector = fmt.Sprintf("workflow_run_id=%s", workflowRunID)

	//List all chaosEngines present in the particular namespace
	chaosEngines, err := dynamicClient.Resource(resourceType).Namespace(namespace).List(context.TODO(), listOption)
	if err != nil {
		return errors.New("failed to list chaosengines: " + err.Error())
	}

	chaosClient, err := GetChaosClient()
	if err != nil {
		fmt.Println(err)
	}

	for _, val := range chaosEngines.Items {
		var (
			crd    *v1alpha1.ChaosEngine
			expRes *v1alpha1.ChaosResult
		)
		crd, err = chaosClient.ChaosEngines(val.GetNamespace()).Get(ctx, val.GetName(), v1.GetOptions{})
		if err != nil {
			fmt.Println("Err:", err)
		}
		if strings.ToLower(string(crd.Status.EngineStatus)) == "completed" {
			expRes, err = chaosClient.ChaosResults(val.GetNamespace()).Get(context.Background(), crd.Name+"-"+crd.Status.Experiments[0].Name, v1.GetOptions{})
			if err != nil {
				fmt.Println(err)
			}
			if strings.ToLower(string(expRes.Status.ExperimentStatus.Verdict)) == "fail" {
				expDetails := types.ExperimentDetails{
					ExperimentName:         crd.Labels["workflow_name"],
					EngineName:             crd.Name,
					FailStep:               expRes.Status.ExperimentStatus.FailStep,
					Phase:                  string(expRes.Status.ExperimentStatus.Phase),
					ProbeSuccessPercentage: expRes.Status.ExperimentStatus.ProbeSuccessPercentage,
					ExpPod:                 crd.Status.Experiments[0].ExpPod,
					RunnerPod:              crd.Status.Experiments[0].Runner,
					Namespace:              crd.Namespace,
				}
				fmt.Println("Hello")
				ticketID, err := notify.NotifyJira(c.JiraURL, expDetails)
				if err != nil {
					fmt.Println(err)
				}
				expLogs := GetLogs(expDetails.ExpPod, expDetails.Namespace)
				runnerLogs := GetLogs(expDetails.RunnerPod, expDetails.Namespace)
				fmt.Println(expLogs + "\n" + runnerLogs)
				err = writeLogsInFile(ticketID, expLogs+"\n"+runnerLogs)
				if err != nil {
					fmt.Println(err)
					return err
				}
			}
		}
	}
	return nil
}

func GetLogs(podName string, namespace string) string {
	ctx := context.TODO()
	conf, err := GetKubeConfig()
	if err != nil {
		return ""
	}

	podLogOpts := corev1.PodLogOptions{}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return ""
	}

	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return ""
	}

	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return ""
	}

	str := buf.String()

	return str
}

func writeLogsInFile(filename string, logs string) error {
	f, err := os.Create(filename + ".txt")
	if err != nil {
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	_, err2 := f.WriteString(logs)

	if err2 != nil {
		return err
	}

	fmt.Println("done")
	return nil
}

func GetKubeConfig() (*rest.Config, error) {
	// Use in-cluster config if kubeconfig path is not specified
	if *KubeConfig == "" {
		return rest.InClusterConfig()
	}
	return clientcmd.BuildConfigFromFlags("", *KubeConfig)
}

func GetDynamicClient() (dynamic.Interface, error) {
	cfg, err := GetKubeConfig()
	if err != nil {
		return nil, err
	}

	// Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return dyn, err
}

func GetChaosClient() (*v1alpha12.LitmuschaosV1alpha1Client, error) {
	cfg, err := GetKubeConfig()
	if err != nil {
		return nil, err
	}
	chaosClient, err := litmusV1alpha1.NewForConfig(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("could not get Chaos ClientSet")
	}
	return chaosClient, nil
}
