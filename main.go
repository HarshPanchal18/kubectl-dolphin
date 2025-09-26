package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/pflag"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	notGiven = "N/A"
	controlPlaneLabel = "node-role.kubernetes.io/control-plane"
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
	-h, --help               Print plugin usage
	-n, --namespace string   Namespace of the pod(s) (default: "default")
	-w, --node string   	 Node name on which pod(s) are scheduled. (Required)
	-b, --batch-size N   	 Delete in batch of N pod.
	-v, --verbose            Show pod name in output

Example:
	$ kubectl dolphin -n data -w worker1
	`)
}

// Method to get pods of a given namespace deployed on a given node.
func deletePodsOnNode(client *kubernetes.Clientset, namespace string, nodename string, batchSize int) {

	pods, _ := client.CoreV1().Pods(namespace).List(
		context.TODO(), v1.ListOptions{
			// Filter based on node-name earlier.
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodename),
		},
	)

	// Empty Pod list
	if len(pods.Items) == 0 {
		fmt.Fprintf(os.Stderr, "No pods found in namespace %v deployed on node %v.\n", namespace, nodename)
		return
	}

	for i := 0; i < len(pods.Items); i += batchSize {

		// Outbound of batch
		end := i + batchSize
		if end > len(pods.Items) {
			end = len(pods.Items)
		}

		// A slice from pod list
		batchPods := pods.Items[i:end]

		// Start deleting pod one by one.
		for _, pod := range batchPods {

			// If -v/--verbose flag is provided...
			if pflag.CommandLine.Changed("verbose") {
				fmt.Println("Deleting pod: ", pod.Name)
			}

			dryRun := []string(nil)
			// If -v/--verbose flag is provided...
			if pflag.CommandLine.Changed("dry-run") {
				dryRun = []string{"All"}
			}

			client.CoreV1().Pods(namespace).Delete(
				context.TODO(), pod.Name, v1.DeleteOptions{ DryRun: dryRun },
			)
		}

		// Add a delay between batches
        time.Sleep(2 * time.Second)
	}
}

func isNodeControlPlane(client *kubernetes.Clientset, node string) bool {

	_node, _ := client.CoreV1().Nodes().Get(context.TODO(), node, v1.GetOptions{})

	for label := range _node.Labels {
		if label == controlPlaneLabel {
			return true
		}
	}

	return false
}

func main() {

	var help bool
	pflag.BoolVarP(&help, "help", "h", false, "Print usage")

	var verbose bool
	pflag.BoolVarP(&verbose, "verbose", "v", false, "Being more informative")

	var dryrun bool
	pflag.BoolVarP(&dryrun, "dry-run", "", false, "Run in a dry-run manner")

	var namespace string
	pflag.StringVarP(&namespace, "namespace", "n", "default", "Namespace of the pod(s)")

	var node string
	pflag.StringVarP(&node, "node", "w", notGiven, "Name of the node on which pods are scheduled")

	var batchSize int
	pflag.IntVarP(&batchSize, "batch-size", "b", 5, "Number of pods to delete per batch")

	pflag.Parse()

	// Print usage
	if help {
		printUsage()
		return
	}

	// If nodename is not provided
	if node == notGiven {
		fmt.Fprintln(os.Stderr, "Error: --node NODE_NAME is required")
        os.Exit(1)
	}

	client, err := getClientset()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting Kubernetes: %v\n", err)
		os.Exit(1)
	}

	// ! Abort if provided node is control-plane.
	if isNodeControlPlane(client, node) {
		fmt.Fprintf(os.Stderr, "Can't perform this action on a control-plane node.\n")
		return
	}

	// Namespace doesn't exist. Abort.
	if _, err = client.CoreV1().Namespaces().Get(context.TODO(),namespace,v1.GetOptions{}) ; err != nil {
		fmt.Fprintf(os.Stderr, "Namespace %v does not exist.", namespace)
		return
	}

	deletePodsOnNode(client, namespace, node, batchSize)
}