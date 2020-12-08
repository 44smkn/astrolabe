package main

type Plan struct {
	PipelineTrigger      PipelineTrigger `yaml:"pipelineTrigger"`
	Cluster              Cluster         `yaml:"cluster"`
	Target               Target          `yaml:"target"`
	CheckIntervalSeconds int             `yaml:"checkIntervalSeconds"`
	Testcases            []Testcase      `yaml:"testcases"`
}

type PipelineTrigger struct {
	Enabled string `yaml:"enabled"`
	Webhook WebhookOnSpinnaker
	SpinCli SpinCli
}

type Cluster struct {
	Name         string `yaml:"name"`
	CertFilePath string `yaml:"certFilePath"`
}

type Target struct {
	Namespace              string            `yaml:"namespace"`
	Kind                   string            `yaml:"kind"`
	LabelSelector          map[string]string `yaml:"labelSelector"`
	CurrentVersionCriteria string            `yaml:"currentVersionCriteria"`
}

type Testcase struct {
	Name   string  `yaml:"name"`
	States []State `yaml:"states"`
}
