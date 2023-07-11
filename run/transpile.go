package graphql

import (
	"fmt"
	"log"
	"path/filepath"

	api_v1 "github.com/vision-cli/api/v1"
	"github.com/vision-cli/vision-plugin-graphql-v1/run/vast_graphql"

	"github.com/vision-cli/common/cases"
	"github.com/vision-cli/common/execute"
	"github.com/vision-cli/common/tmpl"
	"github.com/vision-cli/common/transpiler/model"
)

func Transpile(p *api_v1.PluginPlaceholders, executor execute.Executor, t tmpl.TmplWriter) error {
	targetDir := filepath.Join(p.ProjectRoot, p.ServicesDirectory, p.ServiceNamespace, p.GraphqlServiceName)

	//This will probably be called from the template plugin and may need to be removed from here
	if err := Create(p, executor, t); err != nil {
		log.Fatalf("Failed to create graphql server: %v\n", err)
	}

	//Remove once actual values are being sent
	service := model.Service{
		Name: "projects",
		Enums: []model.Enum{
			{Name: "project-type", Values: []string{"not-assigned", "internal", "billable"}},
		},
		Entities: []model.Entity{
			{
				Name:        "project",
				Persistence: "db",
				Fields: []model.Field{
					{Name: "name", Type: "string", Tag: "db:", IsArray: false, IsNullable: true, IsSearchable: false},
				},
			},
		},
	}
	//Remove once actual values are being sent
	moduleName := "projects"

	if err := vast_graphql.CreateSchemaGraphql(service, targetDir, moduleName); err != nil {
		return fmt.Errorf("Could not create graphql schema: %w", err)
	}

	if err := vast_graphql.CreateGraphqlResolvers(service, p.ProjectRoot, moduleName,
		p.ServiceFqn, moduleName+"/"+cases.Snake(service.Name),
		p.Remote+"/"+p.ProjectName, executor); err != nil {
		return fmt.Errorf("Could not create graphql resolvers: %w", err)
	}

	if err := vast_graphql.UpdateRootResolver(service, p.ProjectRoot, moduleName); err != nil {
		return fmt.Errorf("Could not create graphql schema: %w", err)
	}

	log.Printf("Successfully created graphql server %s\n", p.ProjectName)
	return nil
}
