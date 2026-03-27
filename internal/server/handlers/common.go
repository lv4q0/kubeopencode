// Copyright Contributors to the KubeOpenCode project

// Package handlers implements HTTP handlers for the KubeOpenCode server.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	kubeopenv1alpha1 "github.com/kubeopencode/kubeopencode/api/v1alpha1"
	"github.com/kubeopencode/kubeopencode/internal/server/types"
)

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, err string, message string) {
	writeJSON(w, status, types.ErrorResponse{
		Error:   err,
		Message: message,
		Code:    status,
	})
}

// clientFromContext returns the impersonated client from context or falls back to the default.
func clientFromContext(ctx context.Context, defaultClient client.Client) client.Client {
	if c, ok := ctx.Value(ClientContextKey{}).(client.Client); ok && c != nil {
		return c
	}
	return defaultClient
}

// resolveAgentServerURL looks up the Agent CR and returns its in-cluster server URL.
func resolveAgentServerURL(ctx context.Context, k8sClient client.Client, namespace, agentName string) (string, error) {
	var agent kubeopenv1alpha1.Agent
	if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: agentName}, &agent); err != nil {
		return "", fmt.Errorf("agent not found: %w", err)
	}

	if agent.Spec.ServerConfig == nil {
		return "", fmt.Errorf("agent %q is not in Server mode (no serverConfig)", agentName)
	}

	if agent.Status.ServerStatus == nil || agent.Status.ServerStatus.URL == "" {
		return "", fmt.Errorf("agent %q server is not ready (no server URL in status)", agentName)
	}

	return agent.Status.ServerStatus.URL, nil
}

// normalizeProxyPath extracts and normalizes the wildcard path suffix from a chi route.
func normalizeProxyPath(r *http.Request) string {
	path := chi.URLParam(r, "*")
	if path == "" {
		return "/"
	}
	if path[0] != '/' {
		return "/" + path
	}
	return path
}

// writeResourceOutput writes a Kubernetes resource as JSON or YAML depending on the output query parameter
func writeResourceOutput(w http.ResponseWriter, r *http.Request, statusCode int, obj client.Object, jsonResponse interface{}) {
	output := r.URL.Query().Get("output")
	if output == "yaml" {
		data, err := json.Marshal(obj)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to marshal resource", err.Error())
			return
		}
		yamlData, err := yaml.JSONToYAML(data)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to convert to YAML", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/x-yaml")
		w.WriteHeader(statusCode)
		_, _ = w.Write(yamlData)
		return
	}
	writeJSON(w, statusCode, jsonResponse)
}
