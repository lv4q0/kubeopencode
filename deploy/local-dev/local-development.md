# Local Development Guide

This guide describes how to set up a local development environment for KubeOpenCode using Kind (Kubernetes in Docker).

## Prerequisites

- Docker
- Kind (`brew install kind` on macOS)
- kubectl
- Helm 3.x
- Go 1.25+

## Quick Start

### 1. Create or Use Existing Kind Cluster

Check if you already have a Kind cluster running:

```bash
kind get clusters
```

If you have an existing cluster (e.g., `kind`), you can use it directly. Otherwise, create a new one:

```bash
kind create cluster --name kubeopencode
```

Verify the cluster is running:

```bash
kubectl cluster-info
```

**Note:** The examples below use `--name kubeopencode` for Kind commands. If using an existing cluster with a different name (e.g., `kind`), replace `--name kubeopencode` with your cluster name.

### 2. Build Images

Build all required images:

```bash
# Build the controller image
make docker-build

# Build all agent images (opencode, devbox, attach, etc.)
make agent-build-all
```

Or build individual agent images as needed:

```bash
make agent-build AGENT=opencode    # OpenCode init container (copies /opencode binary)
make agent-build AGENT=devbox      # Executor container (development environment)
make agent-build AGENT=attach      # Attach image (required for Server mode)
```

> **Note:** The `attach` image is required for **Server mode** Agents. If you only use Pod mode, you can skip it. However, `make agent-build-all` is recommended to avoid missing images.

**Note:** The unified kubeopencode image provides both controller and infrastructure utilities:
- `kubeopencode controller`: Kubernetes controller
- `kubeopencode git-init`: Git repository cloning for Git Context
- `kubeopencode save-session`: Workspace persistence for session resume

### 3. Load Images to Kind

Load images into the Kind cluster (required because Kind cannot pull from local Docker):

```bash
# Load controller image
kind load docker-image quay.io/kubeopencode/kubeopencode:latest --name kubeopencode

# Load all agent images
for img in opencode devbox attach; do
  kind load docker-image quay.io/kubeopencode/kubeopencode-agent-${img}:latest --name kubeopencode
done
```

> **Important:** All three agent images must be loaded. Missing the `attach` image will cause Server mode Tasks to fail with `ErrImagePull`.

### 4. Deploy with Helm

```bash
helm upgrade --install kubeopencode ./charts/kubeopencode \
  --namespace kubeopencode-system \
  --create-namespace \
  --set controller.image.pullPolicy=Never \
  --set agent.image.pullPolicy=Never \
  --set server.enabled=true
```

> **Note:** The `server.enabled=true` deploys the UI server. You can omit it if you only need the controller.

### 5. Verify Deployment

Check the controller is running:

```bash
kubectl get pods -n kubeopencode-system
```

Expected output:

```
NAME                                       READY   STATUS    RESTARTS   AGE
kubeopencode-controller-xxxxxxxxx-xxxxx    1/1     Running   0          30s
kubeopencode-server-xxxxxxxxx-xxxxx        1/1     Running   0          30s
```

> **Note:** If `server.enabled=false`, only the controller pod will be present.

Check CRDs are installed:

```bash
kubectl get crds | grep kubeopencode
```

Expected output:

```
agents.kubeopencode.io            <timestamp>
kubeopencodeconfigs.kubeopencode.io   <timestamp>
tasks.kubeopencode.io             <timestamp>
```

Check controller logs:

```bash
kubectl logs -n kubeopencode-system deployment/kubeopencode-controller
```

## UI Server

KubeOpenCode includes a web UI for managing Tasks and viewing Agents. The UI is an "Agent as Application" platform that allows non-technical users to interact with AI agents.

### Access the UI

#### Option 1: Port Forward (Quick Access)

```bash
kubectl port-forward -n kubeopencode-system svc/kubeopencode-server 2746:2746
```

Then open http://localhost:2746 in your browser.

#### Option 2: NodePort (Kind Cluster)

Update Helm values to expose via NodePort:

```bash
helm upgrade kubeopencode ./charts/kubeopencode \
  --namespace kubeopencode-system \
  --set server.enabled=true \
  --set server.service.type=NodePort \
  --set server.service.nodePort=32746
```

Access via: http://localhost:32746

#### Option 3: Ingress

Enable ingress in Helm values:

```bash
helm upgrade kubeopencode ./charts/kubeopencode \
  --namespace kubeopencode-system \
  --set server.enabled=true \
  --set server.ingress.enabled=true \
  --set server.ingress.hosts[0].host=kubeopencode.local \
  --set server.ingress.hosts[0].paths[0].path=/ \
  --set server.ingress.hosts[0].paths[0].pathType=Prefix
```

Add to `/etc/hosts`:
```
127.0.0.1 kubeopencode.local
```

### UI Features

| Feature | Description |
|---------|-------------|
| **Task List** | View all Tasks across namespaces with status filtering |
| **Task Detail** | View Task details, logs (real-time streaming) |
| **Task Create** | Create new Tasks with Agent selection (filtered by namespace permissions) |
| **Agent List** | Browse available Agents with namespace filter |
| **Agent Detail** | View Agent configuration, contexts, credentials |
| **Filtering** | Filter resources by name and Kubernetes label selectors |
| **Pagination** | Server-side pagination for efficient browsing of large resource lists |

#### Resource Filtering

All list pages (Tasks, Agents) support filtering:

- **Name Filter**: Filter resources by name (substring match)
- **Label Selector**: Filter by Kubernetes labels using standard selector syntax (e.g., `app=myapp,env=prod`)

Filters are persisted in the URL as query parameters, making it easy to share filtered views with team members:

```
http://localhost:2746/agents?name=opencode&labels=env%3Dprod
```

#### Pagination

List pages use server-side pagination with 12 items per page. The pagination controls at the bottom of each list show:
- Current page range (e.g., "Showing 1 to 12 of 45")
- Previous/Next navigation buttons

#### Namespace Filtering

The Agent list page includes a namespace selector that allows you to filter Agents by namespace. This helps in multi-tenant environments where different teams have Agents in different namespaces.

#### Agent Availability

When creating a Task, only Agents in the same namespace as the Task are shown. Task and Agent must always be in the same namespace.

### Authentication

The UI uses ServiceAccount token authentication. For local development with port-forward, the server operates in a permissive mode suitable for testing.

For production, enable authentication with RBAC-based filtering:

```bash
helm upgrade kubeopencode ./charts/kubeopencode \
  --namespace kubeopencode-system \
  --set server.enabled=true \
  --set server.authEnabled=true \
  --set server.authAllowAnonymous=false
```

When authentication is enabled:
- Users must provide a Bearer token in the Authorization header
- The token is validated using the Kubernetes TokenReview API
- API requests are executed with user impersonation, respecting Kubernetes RBAC
- Users only see Agents and Tasks they have permission to access

### UI Development

To develop the UI locally with hot-reload:

```bash
# Terminal 1: Run the Go server (API backend)
make run-server

# Terminal 2: Run webpack dev server (frontend with hot-reload)
make ui-dev
```

The webpack dev server runs on http://localhost:3000 and proxies API requests to the Go server on port 2746.

To build the UI for production:

```bash
make ui-build
```

## Iterative Development

### Controller Changes

When you make changes to the controller code:

```bash
# Rebuild the image
make docker-build

# Reload into Kind
kind load docker-image quay.io/kubeopencode/kubeopencode:latest --name kubeopencode

# Restart the deployment to pick up the new image
kubectl rollout restart deployment/kubeopencode-controller -n kubeopencode-system

# Watch the rollout
kubectl rollout status deployment/kubeopencode-controller -n kubeopencode-system
```

Or use the convenience target:

```bash
make e2e-reload
```

### UI Changes

When you make changes to the UI code:

```bash
# Rebuild the UI and docker image
make ui-build
make docker-build

# Reload into Kind
kind load docker-image quay.io/kubeopencode/kubeopencode:latest --name kubeopencode

# Restart the server deployment
kubectl rollout restart deployment/kubeopencode-server -n kubeopencode-system
```

For faster iteration during UI development, use the dev server instead:

```bash
# Run Go server locally (uses kubeconfig for API access)
make run-server

# In another terminal, run webpack dev server with hot-reload
make ui-dev
```

This provides instant feedback without rebuilding Docker images.

## Local Test Environment

For quick testing, use the pre-configured resources in `deploy/local-dev/`:

### Deploy Test Resources

```bash
# First, create secrets.yaml from template
cp deploy/local-dev/secrets.yaml.example deploy/local-dev/secrets.yaml
# Edit secrets.yaml with your real API keys
vim deploy/local-dev/secrets.yaml

# Deploy all resources (namespace, secrets, RBAC, agents)
kubectl apply -k deploy/local-dev/

# Verify the Agent is ready (for Server mode)
kubectl get agent -n test
kubectl get deployment -n test
```

### Resources Created

| Resource | Name | Description |
|----------|------|-------------|
| Namespace | `test` | Isolated namespace for testing |
| Secret | `opencode-credentials` | OpenCode API key |
| Secret | `git-settings` | Git author/committer settings |
| ServiceAccount | `kubeopencode-agent` | Agent service account |
| Role/RoleBinding | `kubeopencode-agent` | RBAC permissions |
| Agent | `server-agent` | Server-mode agent (persistent) |
| Agent | `pod-agent` | Pod-mode agent (per-task) |

### Test Tasks

#### Server Mode Test

```bash
kubectl apply -n test -f - <<EOF
apiVersion: kubeopencode.io/v1alpha1
kind: Task
metadata:
  name: server-test
spec:
  agentRef:
    name: server-agent
  description: "Say hello world"
EOF

# Check status
kubectl get task -n test
kubectl logs -n test server-test-pod -c agent
```

#### Pod Mode Test

```bash
kubectl apply -n test -f - <<EOF
apiVersion: kubeopencode.io/v1alpha1
kind: Task
metadata:
  name: pod-test
spec:
  agentRef:
    name: pod-agent
  description: "What is 2+2?"
EOF

# Check status
kubectl get task -n test
kubectl logs -n test pod-test-pod -c agent
```

#### Concurrent Tasks Test

```bash
for i in 1 2 3; do
  kubectl apply -n test -f - <<EOF
apiVersion: kubeopencode.io/v1alpha1
kind: Task
metadata:
  name: concurrent-$i
spec:
  agentRef:
    name: server-agent
  description: "Count to $i"
EOF
done

# Watch progress
kubectl get task -n test -w
```

### Customization

#### Using Real Secrets

Create a local secrets file (gitignored):

```bash
cp deploy/local-dev/secrets.yaml deploy/local-dev/secrets.local.yaml
# Edit secrets.local.yaml with real values
kubectl apply -f deploy/local-dev/secrets.local.yaml -n test
```

#### Different AI Model

Edit `agent-server.yaml` or `agent-pod.yaml` to change the model:

```yaml
config: |
  {
    "$schema": "https://opencode.ai/config.json",
    "model": "anthropic/claude-sonnet-4-20250514",
    "small_model": "anthropic/claude-haiku-4-20250514"
  }
```

## Cleanup

### Delete Test Resources

```bash
# Delete all tasks
kubectl delete task --all -n test

# Delete all test resources
kubectl delete -k deploy/local-dev/
```

### Uninstall KubeOpenCode

```bash
helm uninstall kubeopencode -n kubeopencode-system
kubectl delete namespace kubeopencode-system
```

### Delete Kind Cluster

```bash
kind delete cluster --name kubeopencode
```

## Debugging Tools

### Reading OpenCode Stream JSON Output

When running Tasks with `--format json`, the output is in stream-json format which can be hard to read. We provide a utility script to format the output:

```bash
# Read from kubectl logs
kubectl logs <pod-name> -n kubeopencode-system | ./hack/opencode-stream-reader.sh

# Read from a saved log file
cat task-output.log | ./hack/opencode-stream-reader.sh
```

The script requires `jq` and converts the JSON stream into human-readable output with colors and formatting.

## Troubleshooting

### Image Pull Errors

If you see `ErrImagePull` or `ImagePullBackOff`, ensure:

1. Images are loaded into Kind: `docker exec kubeopencode-control-plane crictl images | grep kubeopencode`
2. `imagePullPolicy` is set to `Never` in Helm values
3. **Server mode:** The `attach` image (`kubeopencode-agent-attach`) must be loaded. This image is used by Server mode Task Pods to connect to the OpenCode server. Build and load it with:
   ```bash
   make agent-build AGENT=attach
   kind load docker-image quay.io/kubeopencode/kubeopencode-agent-attach:latest --name kubeopencode
   ```

### Controller Not Starting

Check controller logs:

```bash
kubectl logs -n kubeopencode-system deployment/kubeopencode-controller
```

Check events:

```bash
kubectl get events -n kubeopencode-system --sort-by='.lastTimestamp'
```

### OpenCode Web UI Not Loading in Kind Cluster

The embedded OpenCode Web UI requires static assets to be included in the Docker image at build time. By default, local builds skip the OpenCode Web UI build (`BUILD_OPENCODE_UI=false`) because it requires cloning and building the OpenCode project, which needs network access.

**Symptoms:**
- JS module files return `text/html` MIME type errors
- "Failed to load module script" errors in browser console
- CSP violations for external fonts (`r2cdn.perplexity.ai`)

**Root cause:** Without embedded assets, the Web UI falls back to proxying through the OpenCode server pod, which cannot reach `app.opencode.ai` from inside the Kind cluster (no internet access).

**Solution — Download pre-built assets from CDN:**

Run this script to download the OpenCode Web UI assets directly from the CDN and embed them in the image:

```bash
# Download pre-built OpenCode Web UI assets from CDN
DIST_DIR="internal/opencode-app/dist"
rm -rf "$DIST_DIR" && mkdir -p "$DIST_DIR/assets"

# Download index.html and static assets
curl -s https://app.opencode.ai/ -o "$DIST_DIR/index.html"
for path in $(grep -o '"/[^"]*"' "$DIST_DIR/index.html" | tr -d '"' | sort -u); do
  dest="$DIST_DIR$path"
  mkdir -p "$(dirname "$dest")"
  curl -s "https://app.opencode.ai$path" -o "$dest"
done

# Download lazy-loaded JS chunks from main bundle
MAIN_JS=$(ls $DIST_DIR/assets/index-*.js)
grep -o '"\.\/[^"]*\.js"' "$MAIN_JS" | tr -d '"' | sed 's|^\./||' | sort -u | \
  xargs -P 10 -I{} curl -s "https://app.opencode.ai/assets/{}" -o "$DIST_DIR/assets/{}"

echo "Downloaded $(ls $DIST_DIR/assets/ | wc -l) asset files"

# Rebuild Docker image (assets are now embedded via go:embed)
make docker-build
```

The `go:embed` directive in `internal/opencode-app/embed.go` automatically picks up these files. The handler detects embedded assets and serves them locally (self-hosted mode), only proxying API calls to the OpenCode server.

**Alternative — Build from source** (requires bun):

```bash
# If you have the OpenCode source at ../opencode
make opencode-app-build
make docker-build
```

**Note:** In CI/CD (GitHub Actions), `BUILD_OPENCODE_UI=true` is set automatically, which clones and builds OpenCode from source inside the Docker build stage. This issue only affects local Kind clusters.

### CRDs Not Found

Ensure CRDs are installed:

```bash
kubectl get crds | grep kubeopencode
```

If missing, reinstall with Helm or apply manually:

```bash
kubectl apply -f deploy/crds/
```
