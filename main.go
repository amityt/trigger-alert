package main

import (
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
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
	k8sConfig.ExperimentRunID = "4007c84a-5480-4a21-95eb-cdaa6a3da765"
	k8sConfig.Namespace = "litmus"
	err := k8sConfig.GetChaosEngines(k8sConfig.Namespace, k8sConfig.ExperimentRunID)
	if err != nil {
		fmt.Println(err)
	}
}
