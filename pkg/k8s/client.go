// Package k8s provides utilities for interacting with Kubernetes clusters.
package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientaps a Kubernetes clientset with helper methods.
type Client struct {
	Clientset kubernetes.Interface
	Config    *rest.Config
	Namespace string
}

// NewClient creates a new Kubernetes client using in-cluster config or kubeconfig.
func NewClient(namespace string) (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home directory: %w", err)
			}
			kubeconfig = filepath.Join(home, ".kube", "config")
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	if namespace == "" {
		namespace = "default"
	}

	return &Client{
		Clientset: clientset,
		Config:    config,
		Namespace: namespace,
	}, nil
}

// GetClusterContext returns a summary of the current cluster state for use as LLM context.
func (c *Client) GetClusterContext(ctx context.Context) (string, error) {
	var summary string

	// Get nodes
	nodes, err := c.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list nodes: %w", err)
	}
	summary += fmt.Sprintf("Nodes (%d):\n", len(nodes.Items))
	for _, node := range nodes.Items {
		status := "NotReady"
		for _, cond := range node.Status.Conditions {
			if cond.Type == "Ready" && cond.Status == "True" {
				status = "Ready"
				break
			}
		}
		summary += fmt.Sprintf("  - %s (%s)\n", node.Name, status)
	}

	// Get pods in namespace
	pods, err := c.Clientset.CoreV1().Pods(c.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list pods: %w", err)
	}
	summary += fmt.Sprintf("\nPods in namespace '%s' (%d):\n", c.Namespace, len(pods.Items))
	for _, pod := range pods.Items {
		summary += fmt.Sprintf("  - %s (phase: %s)\n", pod.Name, pod.Status.Phase)
	}

	// Get services in namespace
	services, err := c.Clientset.CoreV1().Services(c.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list services: %w", err)
	}
	summary += fmt.Sprintf("\nServices in namespace '%s' (%d):\n", c.Namespace, len(services.Items))
	for _, svc := range services.Items {
		summary += fmt.Sprintf("  - %s (type: %s)\n", svc.Name, svc.Spec.Type)
	}

	return summary, nil
}
