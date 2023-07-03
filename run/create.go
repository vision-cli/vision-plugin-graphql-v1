package graphql

import (
	"bufio"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	api_v1 "github.com/vision-cli/api/v1"

	"github.com/vision-cli/common/execute"
	"github.com/vision-cli/common/file"
	"github.com/vision-cli/common/module"
	"github.com/vision-cli/common/tmpl"
	"github.com/vision-cli/common/workspace"
	"github.com/vision-cli/vision/cli"
	"github.com/vision-cli/vision/config"
	"github.com/vision-cli/vision/flag"
)

const (
	SubCommand            = "create"
	PlatformTemplateDir   = "_templates/platform"
	StandaloneTemplateDir = "_templates/standalone"
)

//go:embed all:_templates/platform
var platformTemplateFiles embed.FS

//go:embed all:_templates/standalone
var standaloneTemplateFiles embed.FS

func Create(p *api_v1.PluginPlaceholders, executor execute.Executor, t tmpl.TmplWriter) error {
	// p = p.NewDefaultServicePlaceholders(cmd.Flags(), projectRoot, config.GraphqlName())
	targetDir := filepath.Join(p.ProjectRoot, config.ServicesDirectory(), p.ServiceNamespace, config.GraphqlName())

	if config.IsDeploymentStandaloneGateway() {
		return fmt.Errorf("Not generating graphql server for standalone rest deployment")
	}

	reader := bufio.NewReader(os.Stdin)
	if file.Exists(targetDir) &&
		!flag.IsForce(pflag.CommandLine) &&
		!cli.Confirmed(reader, "graphql server already exists, overwrite?") {
		return fmt.Errorf("Not overwriting existing graphql server")
	}

	if config.Deployment() == config.DeployStandaloneGraphql {
		if err := tmpl.GenerateFS(standaloneTemplateFiles, StandaloneTemplateDir, targetDir, p, true, t); err != nil {
			log.Fatalf("Failed to generate standalone graphql server: %v\n", err)
		}
	}

	if config.Deployment() == config.DeployPlatform {
		if err := tmpl.GenerateFS(platformTemplateFiles, PlatformTemplateDir, targetDir, p, true, t); err != nil {
			log.Fatalf("Failed to generate platform graphql server: %v\n", err)
		}
	}

	if err := module.Init(targetDir, p.GraphqlFqn, executor); err != nil {
		return fmt.Errorf("Failed to init module: %w", err)
	}

	if err := module.Tidy(targetDir, executor); err != nil {
		return fmt.Errorf("Failed to tidy module: %w", err)
	}

	relativeTargetDir := filepath.Join(config.ServicesDirectory(), p.ServiceNamespace, config.GraphqlName())
	if err := workspace.Use(p.ProjectRoot, relativeTargetDir, executor); err != nil {
		return fmt.Errorf("Failed to add to workspace: %w", err)
	}

	log.Printf("Successfully created graphql server %s\n", p.ProjectName)
	return nil
}
