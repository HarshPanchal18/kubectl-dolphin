package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	notGiven = "N/A"
)

// getClientset - loads kubeconfig and returns a Kubernetes clientset
func getClientset() (*kubernetes.Clientset, error) {
    kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
        return nil, err
    }
    return kubernetes.NewForConfig(config)
}

func printUsage() {
	fmt.Fprintln(os.Stderr,
			`
			DOLPHIN: Delete On-demand Local Pods Hosted In Node.
Usage:
	Delete all pods of a given namespace hosted on a given node.

Syntax:
	kubectl dolphin POD_NAME -n NAMESPACE
	kubectl dolphin -A -n NAMESPACE

Output:
	CONTAINER: IMAGE

Options:
	-A, --all                List images of all pods in the namespace
	-h, --help               Print plugin usage
	-ns, --namespace string   Namespace of the pod(s) (default: "default")
	-nd, --node string   	Node name on which pod(s) are scheduled. (Required)
	-v, --verbose            Show pod name in output

Example:
	$ kubectl dolphin -ns data -nd worker1
	`)
}

// Method to get pods of a given namespace deployed on a given node.
func getPodsOfNamespaceOnNode(client *kubernetes.Clientset, namespace string, nodename string) {

	pods, _ := client.CoreV1().Pods(namespace).List(context.TODO(),v1.ListOptions{})

	for _, pod := range pods.Items {
		if(pod.Spec.NodeName == nodename) {
			fmt.Println(pod.Name)
		}
		// fmt.Println(pod.Spec.NodeName)
	}

}

// func printAllNodeName(client *kubernetes.Clientset) {
// 	nodes, _ := client.CoreV1().Nodes().List(context.TODO(),v1.ListOptions{})

// 	for _, node := range nodes.Items {
// 		fmt.Println(node.Name)
// 	}

// }

func isNodeControlPlane(client *kubernetes.Clientset, node string) bool {

	// node, _ := client.CoreV1().Nodes().Get(context.TODO(), node, v1.GetOptions{})
	_node, _ := client.CoreV1().Nodes().Get(context.TODO(),node,v1.GetOptions{})

	for label := range _node.Labels {
		if label == "node-role.kubernetes.io/control-plane" {
			return true
		}
	}

	return false
}

func main() {

	var help bool
	pflag.BoolVarP(&help, "help", "h", false, "Print usage")

	var namespace string
	pflag.StringVarP(&namespace, "namespace", "n", "default", "Namespace of the pod(s)")

	var node string
	pflag.StringVarP(&node, "node", "w", notGiven, "Name of the node on which pods are scheduled")

	pflag.Parse()

	if help {
		printUsage()
		return
	}

	fmt.Println(node)

	if node == notGiven {
		fmt.Fprintln(os.Stderr, "Error: --node NODE_NAME is required")
		printUsage()
        os.Exit(1)
	}

	client, err := getClientset()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	// args := pflag.Args()

	getPodsOfNamespaceOnNode(client, namespace, node)

	fmt.Println(isNodeControlPlane(client, node))
}