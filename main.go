package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// TODO Read from the KUBECONFIG env var
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path(s) to the kubeconfig file(s)")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path(s) to the kubeconfig file(s)")
	}
	// TODO Support multiple contexts
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// List Submariner-labeled nodes
	selector := labels.SelectorFromSet(labels.Set(map[string]string{"submariner.io/gateway": "true"}))
	labeledNodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d labeled nodes in the cluster\n", len(labeledNodes.Items))
	for _, node := range labeledNodes.Items {
		for _, label := range node.GetLabels() {
			fmt.Printf("Node %s, label %s\n", node.GetName(), label)
		}
	}

	if len(labeledNodes.Items) == 0 {
		// List all nodes and select one
		allNodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d nodes in the cluster\n", len(allNodes.Items))
		allNodeNames := []string{}
		for _, node := range allNodes.Items {
			allNodeNames = append(allNodeNames, node.GetName())
		}
		var qs = []*survey.Question{
			{
				Name: "node",
				Prompt: &survey.Select{
					Message: "Which node should be used as the gateway?",
					Options: allNodeNames},
			},
		}
		answers := struct {
			Node string
		}{}

		err = survey.Ask(qs, &answers)
		if err != nil {
			panic(err.Error())
		}

		// TODO label the node
	}

	// List pods
	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
