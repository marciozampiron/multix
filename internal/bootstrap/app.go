// File: internal/bootstrap/app.go
// Company: Hassan
// Creator: Zamp
// Created: 15/03/2026
// Updated: 15/03/2026
// Purpose: Orchestrates platform startup by injecting required dependencies.

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

// App encapsulates the entire application state and command tree.
type App struct {
	Config           *config.Config
	Logger           logger.Logger
	RootCmd          *cobra.Command
	ProviderRegistry outbound.ProviderRegistry
	SkillRegistry    *domainSkills.Registry
	SkillExecutor    *skills.Executor
}

// BuildApp constructs the core dependencies and prepares the runtime.
func BuildApp() *App {
	cfg := LoadConfig()
	log := logger.New("info")
	rootCmd := root.NewRootCmd()

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
	}
}

// Wire binds the application commands to the Cobra root command.
func Wire(app *App) *cobra.Command {
	handlers := []inbound.CLIHandler{
		cli.NewDoctorHandler(app.RootCmd, app.SkillExecutor),
		cli.NewAuthHandler(app.RootCmd, app.SkillExecutor),
		cli.NewInventoryHandler(app.RootCmd, app.SkillExecutor),
		cli.NewK8sHandler(app.RootCmd, app.SkillExecutor),
		cli.NewAIHandler(app.RootCmd, app.SkillExecutor),
	}

	for _, h := range handlers {
		h.Register()
	}

	root.RegisterVersionCmd(app.RootCmd)
	return app.RootCmd
}
