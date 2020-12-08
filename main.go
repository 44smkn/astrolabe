package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func discontinue(msg string, exitCode int) {
	fmt.Println(msg)
	os.Exit(exitCode)
}

func main() {

	// Check the configuration file directory is passed
	if len(os.Args) < 2 {
		discontinue("Usage: astrolabe [CONFIG FILEPATH]", 1)
	}

	// File Existence Checking
	file := os.Args[1]
	if _, err := os.Stat(file); os.IsNotExist(err) {
		discontinue(fmt.Sprintf("%s does not exist\n", file), 1)
	}

	// Loading the configuration file
	data, err := ioutil.ReadFile(file)
	if err != nil {
		discontinue(fmt.Sprintf("failed to read file. %v\n", err), 1)
	}
	plan := Plan{}
	if err := yaml.Unmarshal([]byte(data), &plan); err != nil {
		discontinue(fmt.Sprintf("failed to parse configuration file. %v\n", err), 1)
	}

	fmt.Printf("%#v\n", plan)

	// Authentication to the cluster
	kubeconfig := plan.Cluster.CertFilePath
	provider, err := NewKubernetesProvider(kubeconfig)
	if err != nil {
		discontinue(err.Error(), 1)
	}

	// Execute each testcase
	for _, testcase := range plan.Testcases {
		// find current active Replicaset
		rsList, err := fetchReplicaSets(provider.Client, plan.Target.Namespace, plan.Target.LabelSelector)
		currentRs, err := findReplicaSet(rsList, plan.Target.CurrentVersionCriteria)
		if err != nil {
			discontinue(err.Error(), 1)
		}
		fmt.Printf("Current Active Replicaset is %v\n", currentRs.Name)
		newestRs, err := findReplicaSet(rsList, "newest")
		if err != nil {
			discontinue(err.Error(), 1)
		}

		// Run the pipeline
		trigger, err := NewTrigger(plan.PipelineTrigger)
		if err != nil {
			discontinue(fmt.Sprintf("failed to load trigger configuration. %v\n", err), 1)
		}
		confirmFunc, err := trigger.Pull()
		if err != nil {
			discontinue(fmt.Sprintf("failed to pull trigger of deploy pipeline. %v\n", err), 1)
		}

		// find new ReplicaSet
		var testTargetRs *v1.ReplicaSet
		for {
			rsList, _ := fetchReplicaSets(provider.Client, plan.Target.Namespace, plan.Target.LabelSelector)
			maybeNewest, _ := findReplicaSet(rsList, "newest")
			if maybeNewest.Name != newestRs.Name {
				testTargetRs = maybeNewest
				break
			}
			time.Sleep(time.Duration(plan.CheckIntervalSeconds) * time.Second)
		}

		// Check the status of the cluster
		for _, state := range testcase.States {
			// Recognize the new replicaSet
			for {

				if ok, err := state.CheckPods(currentRs, testTargetRs); err != nil {
					fmt.Printf("getting status of cluster is failed. %v\n", err)
				} else if ok {
					break
				}

				time.Sleep(time.Duration(plan.CheckIntervalSeconds) * time.Second)
				confirmFunc()
			}
		}
	}
}

func findReplicaSet(replicaSets []v1.ReplicaSet, criteria string) (*v1.ReplicaSet, error) {
	if len(replicaSets) == 0 {
		return nil, errors.New("replicaSets does not exist")
	}
	switch criteria {
	case "largest":
		largest := replicaSets[0]
		for _, rs := range replicaSets {
			if replicas := rs.Spec.Replicas; *replicas > *largest.Spec.Replicas {
				largest = rs
			}
		}
		return &largest, nil
	case "newest":
		newest := replicaSets[0]
		for _, rs := range replicaSets {
			if created := rs.CreationTimestamp.Time; created.After(newest.CreationTimestamp.Time) {
				newest = rs
			}
			return &newest, nil
		}
	}
	return nil, errors.New("specified criteria is not allowed value")
}

func fetchReplicaSets(clientset *kubernetes.Clientset, namespace string, labelSelector map[string]string) ([]v1.ReplicaSet, error) {
	ls := metav1.LabelSelector{MatchLabels: labelSelector}
	rsList, err := clientset.AppsV1().ReplicaSets(namespace).List(metav1.ListOptions{LabelSelector: labels.Set(ls.MatchLabels).String()})
	if err != nil {
		return nil, err
	}
	return rsList.Items, nil
}
