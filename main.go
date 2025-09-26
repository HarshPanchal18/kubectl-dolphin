package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	notGiven          = "N/A"
	controlPlaneLabel = "node-role.kubernetes.io/control-plane"
	RESET 			  = "\033[0m"
	RED 			  = "\033[31m"
	GREEN 			  = "\033[32m"
	BLUE 			  = "\033[34m"
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
	-b, --batch-size N   	 Delete in batch of N pod (default: 5)
	-v, --verbose            Show pod name in output

Example:
	$ kubectl dolphin -n data -w worker1
	`)
}

// Method to get pods of a given namespace deployed on a given node.
func deletePodsOnNode(client *kubernetes.Clientset, namespace string, nodename string, batchSize int) {

	pods, _ := client.CoreV1().Pods(namespace).List(
		context.TODO(), v1.ListOptions{
			// Filter pods based on node-name earlier.
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodename),
		},
	)

	// Empty Pod list. EXIT.
	if len(pods.Items) == 0 {
		fmt.Fprintln(os.Stderr, RED, "No pods found in namespace", namespace, " deployed on node ", nodename, ".", RESET)
		return
	}

	// If -b/--batch-size is provided
	if pflag.CommandLine.Changed("batch-size") {

		// Start deleting pod in a batch.
		for i := 0; i < len(pods.Items); i += batchSize {

			// Handle outbound of batch
			end := min(i + batchSize, len(pods.Items))

			// A slice from pod list
			deletePods(pods.Items[i:end], client, namespace)

			// Add a delay between batches
			// time.Sleep(2 * time.Second)
		}
	} else {
		// Start deleting pod.
		deletePods(pods.Items, client, namespace)
	}
}

func deletePods(pods []corev1.Pod, client *kubernetes.Clientset, namespace string) {

	for _, pod := range pods {
		// If -v/--verbose flag is provided...
		if pflag.CommandLine.Changed("verbose") {
			fmt.Println(BLUE, "Pod: ", pod.Name, "is being deleted!", RESET)
		}

		dryRun := []string(nil)
		// If --dry-run flag is provided...
		if pflag.CommandLine.Changed("dry-run") {
			dryRun = []string{"All"}
		}

		err := client.CoreV1().Pods(namespace).Delete(
				context.TODO(), pod.Name, v1.DeleteOptions{DryRun: dryRun},
			)

		if err != nil {
			fmt.Println(RED, err, RESET)
			return
		}
	}
}

func isNodeControlPlane(client *kubernetes.Clientset, node string) bool {

	_node, err := client.CoreV1().Nodes().Get(context.TODO(), node, v1.GetOptions{})

	// Node doesn't exist
	if err != nil {
		fmt.Println(RED, err, RESET)
		return false
	}

	for label := range _node.Labels {
		if label == controlPlaneLabel {
			return true
		}
	}

	return false
}

func isSystemNamespace(client *kubernetes.Clientset, namespace string) bool {

	systemNamespace := []string{"kube-system", "kube-public", "kube-node-lease"}
	_namespace, _ := client.CoreV1().Namespaces().Get(context.TODO(), namespace, v1.GetOptions{})

	return slices.Contains(systemNamespace, _namespace.Name)
}

func validateOptions(client *kubernetes.Clientset, node string, namespace string, batchSize int) bool {

	// Node-name is not provided
	if node == notGiven {
		fmt.Fprintln(os.Stderr, "Error: --node NODE_NAME is required")
		printUsage()
		return false
	}

	// ! Abort. Node has a control-plane role.
	if isNodeControlPlane(client, node) {
		fmt.Fprintln(os.Stderr, RED, "Can't perform this action on a control-plane node.", RESET)
		return false
	}

	// ! Abort. Namespace is system-defined.
	if isSystemNamespace(client, namespace) {
		fmt.Fprintln(os.Stderr, RED, "Can't perform this action on a system-defined namespace.", RESET)
		return false
	}

	// ! Abort. Namespace doesn't exist.
	if _, err := client.CoreV1().Namespaces().Get(context.TODO(), namespace, v1.GetOptions{}); err != nil {
		fmt.Fprintln(os.Stderr, RED, "Namespace ", namespace, " does not exist.", RESET)
		return false
	}

	if batchSize < 1 {
		fmt.Fprintln(os.Stderr, RED, "Batch size should be more than 1.", RESET)
		return false
	}

	return true
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
	pflag.IntVarP(&batchSize, "batch-size", "b", 1, "Number of pods to delete per batch")

	pflag.Parse()

	// Print usage
	if help {
		printUsage()
		return
	}

	client, err := getClientset()
	if err != nil {
		fmt.Fprintln(os.Stderr, RED, "Error connecting Kubernetes: ", err, RESET)
		os.Exit(1)
	}

	// Validate CLI option entries
	validOptions := validateOptions(client, node, namespace, batchSize)
	if !validOptions { return }

	deletePodsOnNode(client, namespace, node, batchSize)

	fmt.Println(GREEN, "Operation completed successfully! Dolphin is underwater. 🐬", RESET)

}