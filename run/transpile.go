package graphql

import (
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
	service := model.Service{Name: "TestService", Enums: []model.Enum{}, Entities: []model.Entity{}}
	moduleName := "TestModule"

	vast_graphql.CreateSchemaGraphql(service, targetDir, moduleName)
	vast_graphql.CreateGraphqlResolvers(service, p.ProjectRoot, moduleName,
		p.ServiceFqn, moduleName+"/"+cases.Snake(p.ServiceName),
		p.Remote+"/"+p.ProjectName, executor)
	vast_graphql.UpdateRootResolver(service, p.ProjectRoot, moduleName)

	log.Printf("Successfully created graphql server %s\n", p.ProjectName)
	return nil
}
