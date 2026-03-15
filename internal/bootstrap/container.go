package bootstrap

import (
	"multix/cmd/root"
	"multix/internal/adapters/inbound/cli"
	"multix/internal/adapters/outbound/ai/gemini"
	"multix/internal/adapters/outbound/cloud/aws"
	"multix/internal/adapters/outbound/cloud/gcp"
	"multix/internal/adapters/outbound/config"

	"multix/internal/application/ai"
	"multix/internal/application/auth"
	"multix/internal/application/doctor"
	"multix/internal/application/inventory"
	"multix/internal/application/k8s"
	"multix/internal/application/skills" // Executor

	domainSkills "multix/internal/domain/skills" // Registry

	"multix/internal/platform/logger"
	"multix/internal/ports/inbound"
	"multix/internal/ports/outbound"

	"github.com/spf13/cobra"
)

type Container struct {
	ConfigStore outbound.ConfigStore
	Logger      logger.Logger
	RootCmd     *cobra.Command

	// V2: Abstract Factory for multi-cloud resolution during runtime
	ProviderRegistry outbound.ProviderRegistry

	// Agent-ready architecture
	SkillRegistry *domainSkills.Registry
	SkillExecutor *skills.Executor
}

func BuildContainer() *Container {
	rootCmd := root.NewRootCmd()
	cfgStore, _ := config.NewFileStore()
	log := logger.New("info")

	// 1. Initialize Adapters
	awsAdapter := aws.NewAdapter(log)
	gcpAdapter := gcp.NewAdapter(log)
	geminiAdapter := gemini.NewAdapter(log)

	// 2. Build Provider Registry (V2 Factory)
	providers := NewBootstrapRegistry()
	providers.RegisterAuth("aws", awsAdapter)
	providers.RegisterInventory("aws", awsAdapter)
	providers.RegisterK8s("aws", awsAdapter)

	providers.RegisterAuth("gcp", gcpAdapter)
	providers.RegisterInventory("gcp", gcpAdapter)
	providers.RegisterK8s("gcp", gcpAdapter)

	providers.RegisterAI("gemini", geminiAdapter)

	// 3. Register the Universal AI/Agent Tools feeding the Registry
	skillRegistry := domainSkills.NewRegistry()
	skillRegistry.Register(doctor.NewCheckEnvSkill())
	skillRegistry.Register(auth.NewLoginSkill(providers))
	skillRegistry.Register(auth.NewValidateSkill(providers))
	skillRegistry.Register(auth.NewWhoamiSkill(providers))
	skillRegistry.Register(inventory.NewScanSkill(providers))
	skillRegistry.Register(inventory.NewSummarySkill(providers))
	skillRegistry.Register(k8s.NewListClustersSkill(providers))
	skillRegistry.Register(ai.NewExplainSkill(providers))

	executor := skills.NewExecutor(skillRegistry)

	return &Container{
		ConfigStore:      cfgStore,
		Logger:           log,
		RootCmd:          rootCmd,
		ProviderRegistry: providers,
		SkillRegistry:    skillRegistry,
		SkillExecutor:    executor,
	}
}

func Wire(c *Container) *cobra.Command {

	handlers := []inbound.CLIHandler{
		cli.NewDoctorHandler(c.RootCmd, c.SkillExecutor),
		cli.NewAuthHandler(c.RootCmd, c.SkillExecutor),
		cli.NewInventoryHandler(c.RootCmd, c.SkillExecutor),
		cli.NewK8sHandler(c.RootCmd, c.SkillExecutor),
		cli.NewAIHandler(c.RootCmd, c.SkillExecutor),
	}

	for _, h := range handlers {
		h.Register()
	}

	root.RegisterVersionCmd(c.RootCmd)
	return c.RootCmd
}
