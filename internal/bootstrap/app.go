// File: internal/bootstrap/app.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Orchestrates platform startup by composing runtime dependencies.

package bootstrap

import (
	"multix/cmd/root"
	"multix/internal/adapters/inbound/cli"
	"multix/internal/application/skills"
	"multix/internal/domain/config"
	domainSkills "multix/internal/domain/skills"
	"multix/internal/platform/logger"
	"multix/internal/ports/inbound"
	"multix/internal/ports/outbound"

	"github.com/spf13/cobra"
)

// App encapsulates the application runtime dependencies and command tree.
type App struct {
	Config           *config.Config
	Logger           logger.Logger
	RootCmd          *cobra.Command
	ProviderRegistry outbound.ProviderRegistry
	SkillRegistry    *domainSkills.Registry
	SkillExecutor    *skills.Executor
}

// BuildApp constructs the application runtime dependencies.
func BuildApp() (*App, error) {
	cfg := LoadConfig()
	log := logger.New("info")
	rootCmd := root.NewRootCmd()
	applyRuntimeDefaults(rootCmd, cfg)

	providers := BuildProviderRegistry(log)
	skillRegistry := BuildSkillRegistry(providers)
	executor := skills.NewExecutor(skillRegistry)

	return &App{
		Config:           cfg,
		Logger:           log,
		RootCmd:          rootCmd,
		ProviderRegistry: providers,
		SkillRegistry:    skillRegistry,
		SkillExecutor:    executor,
	}, nil
}

// Wire registers inbound CLI handlers and returns the wired root command.
func (a *App) Wire() *cobra.Command {
	handlers := []inbound.CLIHandler{
		cli.NewDoctorHandler(a.RootCmd, a.SkillExecutor),
		cli.NewAuthHandler(a.RootCmd, a.SkillExecutor),
		cli.NewInventoryHandler(a.RootCmd, a.SkillExecutor),
		cli.NewK8sHandler(a.RootCmd, a.SkillExecutor),
		cli.NewAIHandler(a.RootCmd, a.SkillExecutor),
	}

	for _, h := range handlers {
		h.Register()
	}

	a.RootCmd.AddCommand(cli.NewServeCmd(a.Logger, a.SkillRegistry))

	root.RegisterVersionCmd(a.RootCmd)
	return a.RootCmd
}

func applyRuntimeDefaults(rootCmd *cobra.Command, cfg *config.Config) {
	_ = rootCmd.PersistentFlags().Set("provider", cfg.DefaultCloudProvider)
	_ = rootCmd.PersistentFlags().Set("output", cfg.DefaultOutputMode)
}
