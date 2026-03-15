# Multix CLI (Skills-First & Agent-Ready)

**Multix** is an Enterprise-Grade Multi-cloud CLI platform engineered for DevOps, Kubernetes, and Generative AI integration.

The architecture strictly adheres to a **Skills-First Capability System** built atop Go 1.25+, Cobra, and a clean Hexagonal/DDD-lite topology.

## 🧠 Architectural Philosophy

Unlike traditional tools mapped directly to Cloud Provider endpoints, Multix views the system through **Reusable Universal Skills**. 

The fundamental rule of the platform: 
> **Agent thinks and orchestrates / Skill executes.**

### The Universal Contract
Every business use case (a "Skill") implements a JSON-capable interface (`domain/skills/Skill`). This allows the *exact same underlying Go binary logic* to be invoked seamlessly by:
1. **The CLI Terminal** (via Cobra handlers translating flags to JSON)
2. **AI Agents** (via LLM function calling and Tool-Usage features on Gemini/OpenAI)
3. **External APIs**

## 📂 Project Layout

```text
multix/
├── cmd/
│   ├── multix/                 # The main entrypoint, minimal and clean
│   └── root/                   # Base Cobra CLI root commands
├── internal/
│   ├── domain/                 # Core domain models and the Universal Skill Contract
│   │   ├── skills/             # Registry, Executor, and interface definitions
│   │   └── {auth, k8s, ...}
│   ├── application/            # The actual Go logic wrapping business capabilities
│   ├── ports/
│   │   ├── inbound/            # Interfaces for CLI Handlers and Agent Tool Handlers
│   │   └── outbound/           # Deeply segregated required interfaces (Auth, K8s, Inventory)
│   ├── adapters/
│   │   ├── inbound/cli/        # Cobra CLI Handlers
│   │   └── outbound/           # Provider implementations (aws, gcp, gemini, config)
│   ├── bootstrap/              # High-performance static dependency injection (Wiring)
│   └── platform/               # Transversal cross-cutting concerns (logger, etc.)
├── Makefile                    # Developer workflow automation
└── README.md
```

## 🛠️ Quick Start

**Prerequisites:** Go 1.25+

```bash
# Clone the repository and fetch dependencies
go mod tidy

# Build the CLI via the Makefile
make build

# Verify installation
./build/multix version
```

### Running CLI Commands

```bash
# Check your environment capabilities
./build/multix doctor

# Authenticate against a cloud provider
./build/multix auth login --provider aws

# Discover assets across your estate
./build/multix inventory list --service compute

# Fetch your managed Kubernetes clusters
./build/multix k8s clusters

# Ask the embedded AI Agent to explain concepts
./build/multix ai explain "CrashLoopBackOff"
```

## 📐 Extending the Application

To add a new skill to the platform:
1. **Define the Skill**: Create a struct in `internal/application/your_domain/` that implements `skills.Skill`. Define its name, description, parameters schema, and execute block.
2. **Register It**: Add it to `internal/bootstrap/container.go` inside the `SkillRegistry`.
3. **Expose It**: Add a `cobra.Command` inside `internal/adapters/inbound/cli/` that passes flags into the executor. Once registered, it is automatically exposed to the Agent Tooling module.

## 🚀 Roadmap (MVP -> Beta -> GA)

### Phase 1: MVP (Complete)
- Core Foundation: Extensible project layout with DDD Lite, Ports & Adapters, and explicit wiring.
- Core Skills: `doctor.run`, `auth.validate`, `auth.whoami`.
- Cloud Assets: `inventory.scan`, `inventory.summary`, `k8s.list_clusters`.
- Agent-Ready Mechanism: Input schemas mapping to LLM payloads.

### Phase 2: Beta
- Provider Enhancements: Real implementations for GCP, Azure, and OCI interfaces.
- Generative AI Integration: Real payload exchange with Gemini/OpenAI using the Agent Handlers for autonomous operations (`ai.generate_terraform`).
- Plugin Capability: Expose dynamic loading of pre-compiled binaries via RPC or standard binary invoking.

### Phase 3: GA (General Availability)
- Cost Estimation: Predictive analytics routines (`cost.estimate`).
- Advanced Security: Code-level static security and compliance assessments (`security.assess`).
