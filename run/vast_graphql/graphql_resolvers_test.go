package vast_graphql

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vision-cli/common/transpiler/model"
)

func TestVerbFunctionWrapperReturnsWrapper(t *testing.T) {
	method, err := verbFunctionWrapper("Project", "Projects", "Client", "Get", nil)
	assert.NoError(t, err)

	expected := `
func (r *ProjectsClientResolver) ProjectsGetProject(args ProjectsGetProjectRequest) (*ProjectsProject, error) {
req := PbToGoStruct[ProjectsGetProjectRequest, pb.GetProjectRequest](args, false)
result, err := r.srv.GetProject(nil, &req)
if err != nil {
	return nil, err
}
resp := PbToGoStruct[pb.GetProjectResponse, ProjectsProject](*result, true)
return &resp, nil
}
`
	assert.Equal(t, expected, method)
}

func TestAddGoTypesReturnsCRUDMessageObjects(t *testing.T) {
	s := model.Service{
		Name: "Projects",
		Entities: []model.Entity{
			{
				Name: "Project",
				Fields: []model.Field{
					{
						Name: "Name",
						Type: "string",
					},
					{
						Name: "Description",
						Type: "string",
					},
				},
			},
			{
				Name: "Milestone",
				Fields: []model.Field{
					{
						Name: "Name",
						Type: "string",
					},
					{
						Name: "Date",
						Type: "string",
					},
				},
			},
		},
	}
	astGoModels := NewBlankAst("resolvers", s)
	astGoModels.addGoTypes(s.Entities, "projects")

	expected := `package resolvers

import "github.com/google/uuid"

type ProjectsGetProjectRequest struct {
	ID graphql.ID` + ` ` + `
}
type ProjectsProject struct {
	ID		graphql.ID` + "\t" + `
	Name		string` + "\t\t" + `
	Description	string` + "\t\t" + `
}
type ProjectsCreateProjectRequest struct {
	Name		string` + "\t" + `
	Description	string` + "\t" + `
}
type ProjectsCreateProjectResponse struct {
	ID graphql.ID` + " " + `
}
type ProjectsUpdateProjectRequest struct {
	ID		graphql.ID` + "\t" + `
	Name		string` + "\t\t" + `
	Description	string` + "\t\t" + `
}
type ProjectsUpdateProjectResponse struct {
	Message string` + " " + `
}
type ProjectsDeleteProjectRequest struct {
	ID graphql.ID` + " " + `
}
type ProjectsDeleteProjectResponse struct {
	Message string` + " " + `
}
type ProjectsGetMilestoneRequest struct {
	ID graphql.ID` + " " + `
}
type ProjectsMilestone struct {
	ID	graphql.ID` + "\t" + `
	Name	string` + "\t\t" + `
	Date	string` + "\t\t" + `
}
type ProjectsCreateMilestoneRequest struct {
	Name	string` + "\t" + `
	Date	string` + "\t" + `
}
type ProjectsCreateMilestoneResponse struct {
	ID graphql.ID` + " " + `
}
type ProjectsUpdateMilestoneRequest struct {
	ID	graphql.ID` + "\t" + `
	Name	string` + "\t\t" + `
	Date	string` + "\t\t" + `
}
type ProjectsUpdateMilestoneResponse struct {
	Message string` + " " + `
}
type ProjectsDeleteMilestoneRequest struct {
	ID graphql.ID` + " " + `
}
type ProjectsDeleteMilestoneResponse struct {
	Message string` + " " + `
}
`
	assert.Equal(t, expected, astGoModels.String())
}

func TestResolverStruct(t *testing.T) {
	code, err := resolverStruct("Projects", "Client")
	assert.NoError(t, err)

	expected := `
type ProjectsClientResolver struct {
	srv *server.Server
}

func NewProjectsClientResolver() ProjectsClientResolver {
	conf := config.MustLoadConfig()
	srv := server.MustLoadServer(conf)

	return ProjectsClientResolver{
		srv: srv,
	}
}
`
	assert.Equal(t, expected, code)
}
