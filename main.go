package main

import (
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"runtime"
	"trigger-alert/k8s"
)

func init() {
	logrus.Info("Go Version: ", runtime.Version())
	logrus.Info("Go OS/Arch: ", runtime.GOOS, "/", runtime.GOARCH)

	k8s.KubeConfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()
}

func main() {
	k8sConfig := k8s.LoadEnv()
	k8sConfig.ExperimentRunID = "922a12dc-0045-475c-96da-a53441a02c83"
	k8sConfig.Namespace = "litmus"
	err := k8sConfig.GetChaosEngines(k8sConfig.Namespace, k8sConfig.ExperimentRunID)
	if err != nil {
		fmt.Println(err)
	}
}
