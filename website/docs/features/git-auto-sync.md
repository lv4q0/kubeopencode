# Git Auto-Sync

Git contexts can be configured to automatically sync with the remote repository.
This enables GitOps workflows where pushing to Git triggers Agent updates.

## Sync Policies

| Policy | Mechanism | Pod Restart | Best For |
|--------|-----------|-------------|----------|
| **HotReload** (default) | Sidecar `git-sync` pulls in-place | No | Prompts, docs, context files |
| **Rollout** | Controller detects change, triggers rolling update | Yes | Workspace root, configs loaded at startup |

## HotReload Example

```yaml
spec:
  contexts:
  - name: team-prompts
    type: Git
    git:
      repository: https://github.com/org/prompts.git
      ref: main
      sync:
        enabled: true
        interval: 5m        # default
        policy: HotReload   # default
    mountPath: prompts/
```

## Rollout Example

```yaml
spec:
  contexts:
  - name: agent-config
    type: Git
    git:
      repository: https://github.com/org/agent-config.git
      ref: main
      sync:
        enabled: true
        interval: 10m
        policy: Rollout
    mountPath: "."
```

## Task Protection (Rollout Policy)

When the controller detects a remote change with Rollout policy, it checks for active Tasks:
- **No active Tasks**: proceeds with rolling update immediately
- **Active Tasks exist**: sets `GitSyncPending` condition and waits
- **Safety timeout (1 hour)**: forces rollout even if Tasks are still running

New Tasks are **not blocked** during `GitSyncPending` — only the Deployment rollout is delayed.

## Sync Status

Agent status tracks the sync state:

```yaml
status:
  gitSyncStatuses:
  - name: team-prompts
    commitHash: "a1b2c3d4e5f6..."
    lastSynced: "2026-04-02T10:30:00Z"
```
