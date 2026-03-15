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

type App struct {
	Config           *config.Config
	Logger           logger.Logger
	RootCmd          *cobra.Command
	ProviderRegistry outbound.ProviderRegistry
	SkillRegistry    *domainSkills.Registry
	SkillExecutor    *skills.Executor
}

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
