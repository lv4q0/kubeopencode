# Skills

Skills are reusable AI agent capabilities defined as SKILL.md files (Markdown with YAML frontmatter). KubeOpenCode supports referencing external skills from Git repositories, allowing teams to share and reuse skills across Agents.

Skills are semantically different from Contexts:
- **Contexts** provide knowledge ("what the agent knows")
- **Skills** provide capabilities ("what the agent can do")

## Basic Usage

```yaml
apiVersion: kubeopencode.io/v1alpha1
kind: Agent
metadata:
  name: skilled-agent
spec:
  workspaceDir: /workspace
  serviceAccountName: kubeopencode-agent
  skills:
  - name: official-skills
    git:
      repository: https://github.com/anthropics/skills.git
      ref: main
      path: skills/
      names:
      - frontend-design
      - webapp-testing
```

## Select Specific Skills

Use the `names` field to include only specific skills from a repository. If omitted, all skills under `path` are included:

```yaml
skills:
- name: team-skills
  git:
    repository: https://github.com/my-org/ai-skills.git
    path: engineering/
    names:
    - code-review
    - testing-strategy
```

## Private Repositories

Use `secretRef` for authentication (same Secret format as Git contexts):

```yaml
skills:
- name: internal-skills
  git:
    repository: https://github.com/my-org/private-skills.git
    ref: v2.0.0
    secretRef:
      name: github-pat
```

## How It Works

1. The controller clones skill Git repositories via `git-init` init containers
2. Skills are mounted at `/skills/{source-name}/` in the agent pod
3. The controller auto-injects `skills.paths` into `opencode.json`
4. OpenCode discovers SKILL.md files and makes them available as slash commands

## Skills in Templates

Skills can be defined in AgentTemplates and inherited by Agents. Agent-level skills replace template-level skills (same merge strategy as contexts):

```yaml
apiVersion: kubeopencode.io/v1alpha1
kind: AgentTemplate
metadata:
  name: base-template
spec:
  workspaceDir: /workspace
  serviceAccountName: kubeopencode-agent
  skills:
  - name: org-standards
    git:
      repository: https://github.com/my-org/standards-skills.git
```

## Field Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | string | (required) | Unique identifier for this skill source |
| `git.repository` | string | (required) | Git URL (https://, http://, or git@) |
| `git.ref` | string | HEAD | Branch, tag, or commit SHA |
| `git.path` | string | (root) | Base directory in repo where skills are located |
| `git.names` | []string | (all) | Specific skill directories to include |
| `git.depth` | int | 1 | Clone depth (1=shallow, 0=full) |
| `git.recurseSubmodules` | bool | false | Clone submodules recursively |
| `git.secretRef.name` | string | - | Secret for Git authentication |
