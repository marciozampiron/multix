You are working inside the Multix repository.

Upgrade the current scaffold from v0.1 to v0.2 while preserving the existing skills-first and agent-ready architecture.

IMPORTANT:
- Make minimal, coherent changes.
- Do not rewrite the whole repository.
- Keep the current structure and improve it incrementally.
- Respect GEMINI.md and all .agent/*.md rules.

## Goals for v0.2

Implement the following improvements:

### 1. Provider Registry
Replace "default provider" assumptions with explicit registries.

Create a provider registry layer that supports:
- cloud auth providers
- cloud inventory providers
- kubernetes providers
- AI providers

Required methods:
- GetCloudAuthProvider(name string)
- GetCloudInventoryProvider(name string)
- GetKubernetesProvider(name string)
- GetAIProvider(name string)

Support at least:
- aws
- gcp
- gemini

### 2. Command Flags
Add `--provider` flag to relevant commands:
- auth validate
- auth whoami
- inventory scan
- inventory summary
- k8s list-clusters
- ai explain
- ai terraform

Provider defaults:
- cloud default: aws
- ai default: gemini

### 3. Output Modes
Add `--output` flag with support for:
- json (default)
- table

Create a rendering helper in:
- `internal/adapters/inbound/cli/render.go`

Keep rendering out of use cases and out of skill logic.

### 4. Config Bootstrap
Create an initial config bootstrap:
- `internal/bootstrap/config.go`
- `internal/domain/config/config.go`

Support:
- app name
- version
- default cloud provider
- default ai provider
- default output mode

Use a simple in-memory/default config for now.
Do NOT add Viper yet.

### 5. Skill Input Support
Update skill execution so that provider can be passed in input maps where relevant.

Example:
- `auth.validate` input can include `provider`
- `inventory.scan` input can include `provider`
- `k8s.list_clusters` input can include `provider`
- `ai.explain` input can include `provider`

### 6. Registry-Aware Skill Wiring
Refactor `bootstrap/skills.go` so that skills can dynamically resolve providers based on input, using the provider registry, rather than using only default providers.

### 7. Agent Adapter Base
Create:
- `internal/adapters/inbound/agent/tool_adapter.go`
- `internal/adapters/inbound/agent/manifest.go`

This layer should:
- expose all registered skills as tool-like manifests
- provide a method to execute a skill by name and input
- stay decoupled from Gemini/OpenAI specifics
- prepare for future MCP integration

### 8. Tests
Add or update tests for:
- provider registry
- provider resolution by name
- skill execution using explicit provider
- output renderer (basic)
- agent adapter manifest listing

### 9. README Update
Update README to include:
- provider flags
- output flags
- examples for aws/gcp/gemini
- mention agent adapter base

## Constraints
- Keep idiomatic Go
- Keep explicit wiring
- Do not introduce heavy dependencies
- Do not use Viper yet
- Do not use native Go plugin loading
- Do not over-engineer
- Prefer small, focused diffs

## Delivery format
Return:
1. files to change
2. files to add
3. updated code for each file
4. short rationale for each change
5. final sanity-check notes