package types

type ExperimentDetails struct {
	ExperimentName         string
	EngineName             string
	FailStep               string
	Phase                  string
	ProbeSuccessPercentage string
	ExpPod                 string
	RunnerPod              string
	Namespace              string
}
