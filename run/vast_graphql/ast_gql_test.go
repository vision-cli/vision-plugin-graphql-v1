package vast_graphql

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vision-cli/common/transpiler/model"
)

func TestAddTypeAddsTypeAndQueryMutations(t *testing.T) {
	fields := []model.Field{
		{
			Name:       "id",
			Type:       "id",
			Tag:        "`gorm:\"column:id;primarykey;\"`",
			IsArray:    false,
			IsNullable: false,
		},
		{
			Name:       "name_of_project",
			Type:       "string",
			Tag:        "`gorm:\"column:name;\"`",
			IsArray:    false,
			IsNullable: false,
		},
		{
			Name:       "not_mandatory",
			Type:       "integer",
			Tag:        "`gorm:\"column:name;\"`",
			IsArray:    false,
			IsNullable: true,
		},
		{
			Name:       "array_field",
			Type:       "unsigned",
			Tag:        "`gorm:\"column:name;\"`",
			IsArray:    true,
			IsNullable: false,
		},
	}
	expectedStr := `type TestProject {
  id: ID!
  nameOfProject: String!
  notMandatory: Int
  arrayField: []Int!
}

type MultiTestProject {
  testProjects: [TestProject]!
  pagination: Pagination
}

extend type Query {
  testGetProject(id: ID!): TestProject
  testListProjects(pagination: PaginationInput, searchBy: String): MultiTestProject!
}

extend type Mutation {
  testCreateProject(nameOfProject: String!, notMandatory: Int, arrayField: []Int!): ID!
  testUpdateProject(id: ID!, nameOfProject: String!, notMandatory: Int, arrayField: []Int!): String!
  testDeleteProject(id: ID!): String!
}
`
	a := NewGqlAst("test")
	a.AddType("Project", fields, false)
	assert.Equal(t, expectedStr, a.String())
}
