# kubectl-dolphin

Delete all pods of a given namespace hosted on a given node.

## Usage

```bash
DOLPHIN: Delete On-demand Local Pods Hosted In a Node.

Usage:
    Delete all pods of a given namespace hosted on a given node.
    Very useful when a node is going in maintenance and you need to schedule pods on other nodes.

Syntax:
    kubectl dolphin POD_NAME -n NAMESPACE
    kubectl dolphin -A -n NAMESPACE

Options:
    -h, --help               Print plugin usage.
    -n, --namespace string   Namespace of the pod(s) (default: "default").
    -w, --node string     Node name on which pod(s) are scheduled (Required).
    -b, --batch-size N     Delete in batch of N pod (default: 0).
    -i, --interval N  Wait for N seconds before deleting next batch (default & minimum: 0s). Works along with --batch-size.
    --delete-daemons         Delete daemonset too.
    --dry-run                Run in dry-run manner.
    -v, --verbose            Show pod name in output.

Example:
    $ kubectl dolphin -n data -w worker1
    Operation completed successfully! Dolphin is underwater. 🐬

    $ kubectl dolphin --node kube-worker2 --namespace web --batch-size 2 --dry-run -i 3s
    Operation completed successfully! Dolphin is underwater. 🐬

    $ kubectl dolphin --node kube-worker2 --namespace web -i 3s
    Operation completed successfully! Dolphin is underwater. 🐬

    $ kubectl dolphin --node kube-worker2 --namespace webi -i 3s
    Namespace 'webi' does not exist.

    $ kubectl dolphin --node kube-worker2 --namespace web -i 3s --batch-size -3
    Batch size must be ⩾ 1.
    See 'kubectl dolphin -h' for help and examples.

    $ kubectl dolphin --node kube-control-plane --namespace web
    Can't perform this action on a control-plane node.

    $ kubectl dolphin --node kube-worker3 --namespace web
    nodes "kind-kube-worker3" not found.
```
