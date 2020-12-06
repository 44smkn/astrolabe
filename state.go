package main

import (
	"errors"
	"strconv"
	"strings"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type State struct {
	Name string `yaml:"name"`
	Pod  struct {
		New     string `yaml:"new"`
		Current string `yaml:"current"`
	} `yaml:"pod,omitempty"`
	Endpoints []Endpoint `yaml:"endpoints,omitempty"`
}

type Endpoint struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Count     string `yaml:"count"`
}

func (s *State) Check(clientset *kubernetes.Clientset, current, new *v1.ReplicaSet) (bool, error) {
	if ok, err := s.CheckPods(current, new); err != nil {
		return false, err
	} else if !ok {
		return false, nil
	}

	// CheckEndpoint
	for _, endpoint := range s.Endpoints {
		endpoint.CheckEndpoint(clientset)
	}

	return true, nil
}

func (s *State) CheckPods(current, new *v1.ReplicaSet) (bool, error) {
	// new pod counts check
	if cond := s.Pod.New; cond != "" {
		if ok, err := checkCount(int(*new.Spec.Replicas), cond); err != nil {
			return false, err
		} else if !ok {
			return false, nil
		}
	}

	// current pods count check
	if cond := s.Pod.New; cond != "" {
		if ok, err := checkCount(int(*current.Spec.Replicas), cond); err != nil {
			return false, err
		} else if !ok {
			return false, nil
		}
	}

	return true, nil
}

func (e *Endpoint) CheckEndpoint(clientset *kubernetes.Clientset) (bool, error) {
	ep, err := clientset.CoreV1().Endpoints(e.Namespace).Get(e.Name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	actual := len(ep.Subsets[0].Addresses)

	return checkCount(actual, e.Count)
}

func checkCount(actural int, cond string) (bool, error) {

	equalCount, err := strconv.Atoi(cond)
	if err == nil {
		return actural == equalCount, nil
	}
	if strings.HasPrefix(cond, ">") {
		count, _ := strconv.Atoi(strings.TrimPrefix(cond, ">"))
		return actural > count, nil
	}
	if strings.HasPrefix(cond, "<") {
		count, _ := strconv.Atoi(strings.TrimPrefix(cond, "<"))
		return actural < count, nil
	}
	return false, errors.New("The input string does not match the format")
}
