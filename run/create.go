package graphql

import (
	"bufio"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	api_v1 "github.com/vision-cli/api/v1"

	"github.com/vision-cli/common/execute"
	"github.com/vision-cli/common/file"
	"github.com/vision-cli/common/module"
	"github.com/vision-cli/common/tmpl"
	"github.com/vision-cli/common/workspace"
	"github.com/vision-cli/vision/cli"
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
	targetDir := filepath.Join(p.ProjectRoot, p.ServicesDirectory, p.ServiceNamespace, p.GraphqlServiceName)

	if p.Deployment == "standalone-gateway" {
		return fmt.Errorf("Not generating graphql server for standalone rest deployment")
	}

	reader := bufio.NewReader(os.Stdin)
	if file.Exists(targetDir) &&
		// !flag.IsForce(pflag.CommandLine) &&  // Needs to be added back in. Waiting for flags to be passed from api
		!cli.Confirmed(reader, "graphql server already exists, overwrite?") {
		return fmt.Errorf("Not overwriting existing graphql server")
	}

	if p.Deployment == "standalone-graphql" {
		if err := tmpl.GenerateFS(standaloneTemplateFiles, StandaloneTemplateDir, targetDir, p, true, t); err != nil {
			log.Fatalf("Failed to generate standalone graphql server: %v\n", err)
		}
	}

	if p.Deployment == "platform" {
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

	relativeTargetDir := filepath.Join(p.ServicesDirectory, p.ServiceNamespace, p.GraphqlServiceName)
	if err := workspace.Use(p.ProjectRoot, relativeTargetDir, executor); err != nil {
		return fmt.Errorf("Failed to add to workspace: %w", err)
	}

	log.Printf("Successfully created graphql server %s\n", p.ProjectName)
	return nil
}
