package vast_graphql

import (
	"io/ioutil"
	"path/filepath"

	"github.com/graphql-go/graphql/language/ast"

	"github.com/vision-cli/common/cases"
	"github.com/vision-cli/common/transpiler/model"
)

func CreateSchemaGraphql(service model.Service, targetDir string, moduleName string) {
	astGql := NewGqlAst(moduleName)
	astGql.addListTypes()
	for _, entity := range service.Entities {
		// add filter input
		isFilterStructRequired := astGql.addFilterTypes(entity)
		// graphql types
		astGql.AddType(cases.Pascal(entity.Name), append([]model.Field{
			{Name: "id", Type: "id", Tag: "`gorm:\"column:id;primarykey;\"`", IsArray: false, IsNullable: false, IsSearchable: false},
		}, entity.Fields...), isFilterStructRequired)
	}
	path := filepath.Join(targetDir, "proto", "schema.graphql")
	err := ioutil.WriteFile(path, []byte(astGql.String()), WriteMode)
	if err != nil {
		panic(err)
	}
}

func (a *AstGql) addFilterTypes(entity model.Entity) bool {
	//TODO add filter types as booleans so they aren't mandatory
	filterFields := []*ast.InputValueDefinition{}
	for _, field := range entity.Fields {
		if field.IsSearchable && (field.Type == model.TypeBoolean || field.Type == model.TypeEnum || field.Type == model.TypeId) {
			filterFields = append(filterFields,
				ast.NewInputValueDefinition(&ast.InputValueDefinition{
					Name: ast.NewName(&ast.Name{
						Value: cases.Camel(field.Name),
					}),
					Directives: []*ast.Directive{},
					Type: ast.NewNamed(&ast.Named{
						Name: ast.NewName(&ast.Name{
							Value: field.FilterGqlType(),
						}),
					}),
				}),
			)
			if field.Type == model.TypeEnum {
				enumFields := []*ast.InputValueDefinition{
					ast.NewInputValueDefinition(&ast.InputValueDefinition{
						Name: ast.NewName(&ast.Name{
							Value: "nullable",
						}),
						Directives: []*ast.Directive{},
						Type: ast.NewNamed(&ast.Named{
							Name: ast.NewName(&ast.Name{
								Value: model.TypeGqlString,
							}),
						}),
					}),
					ast.NewInputValueDefinition(&ast.InputValueDefinition{
						Name: ast.NewName(&ast.Name{
							Value: cases.Camel(field.Name),
						}),
						Directives: []*ast.Directive{},
						Type: ast.NewNamed(&ast.Named{
							Name: ast.NewName(&ast.Name{
								Value: model.TypeGqlString,
							}),
						}),
					}),
				}
				enumDef := ast.NewInputObjectDefinition(&ast.InputObjectDefinition{
					Name: ast.NewName(&ast.Name{
						Value: cases.Pascal(field.Name),
					}),
					Fields: enumFields,
				})
				a.file.Definitions = append(a.file.Definitions, enumDef)
			}
		}
	}
	if len(filterFields) == 0 {
		return false
	}
	filterInput := ast.NewInputObjectDefinition(&ast.InputObjectDefinition{
		Name: ast.NewName(&ast.Name{
			Value: "Filter" + cases.Pascal(entity.Name),
		}),
		Fields: filterFields,
	})
	a.file.Definitions = append(a.file.Definitions, filterInput)
	return true
}

func (a *AstGql) addListTypes() {
	paginationDefinition := ast.NewObjectDefinition(&ast.ObjectDefinition{
		Name: ast.NewName(&ast.Name{
			Value: "Pagination",
		}),
		Directives: []*ast.Directive{},
		Interfaces: []*ast.Named{},
		Fields: []*ast.FieldDefinition{
			ast.NewFieldDefinition(&ast.FieldDefinition{
				Name: ast.NewName(&ast.Name{
					Value: "Limit",
				}),
				Arguments:  []*ast.InputValueDefinition{},
				Directives: []*ast.Directive{},
				Type: ast.NewNamed(&ast.Named{
					Name: ast.NewName(&ast.Name{
						Value: "Int",
					}),
				}),
			}),
			ast.NewFieldDefinition(&ast.FieldDefinition{
				Name: ast.NewName(&ast.Name{
					Value: "Offset",
				}),
				Arguments:  []*ast.InputValueDefinition{},
				Directives: []*ast.Directive{},
				Type: ast.NewNamed(&ast.Named{
					Name: ast.NewName(&ast.Name{
						Value: "Int",
					}),
				}),
			}),
			ast.NewFieldDefinition(&ast.FieldDefinition{
				Name: ast.NewName(&ast.Name{
					Value: "Total",
				}),
				Arguments:  []*ast.InputValueDefinition{},
				Directives: []*ast.Directive{},
				Type: ast.NewNamed(&ast.Named{
					Name: ast.NewName(&ast.Name{
						Value: "Int",
					}),
				}),
			}),
		},
	})
	a.file.Definitions = append(a.file.Definitions, paginationDefinition)

	paginationInput := ast.NewInputObjectDefinition(&ast.InputObjectDefinition{
		Name: ast.NewName(&ast.Name{
			Value: "PaginationInput",
		}),
		Fields: []*ast.InputValueDefinition{
			ast.NewInputValueDefinition(&ast.InputValueDefinition{
				Name: ast.NewName(&ast.Name{
					Value: "Limit",
				}),
				Directives: []*ast.Directive{},
				Type: ast.NewNamed(&ast.Named{
					Name: ast.NewName(&ast.Name{
						Value: "Int",
					}),
				}),
			}),
			ast.NewInputValueDefinition(&ast.InputValueDefinition{
				Name: ast.NewName(&ast.Name{
					Value: "Offset",
				}),
				Directives: []*ast.Directive{},
				Type: ast.NewNamed(&ast.Named{
					Name: ast.NewName(&ast.Name{
						Value: "Int",
					}),
				}),
			}),
		},
	})
	a.file.Definitions = append(a.file.Definitions, paginationInput)

	// searchInput := ast.NewInputObjectDefinition(&ast.InputObjectDefinition{
	// 	Name: ast.NewName(&ast.Name{
	// 		Value: "SearchInput",
	// 	}),
	// 	Fields: []*ast.InputValueDefinition{
	// 		ast.NewInputValueDefinition(&ast.InputValueDefinition{
	// 			Name: ast.NewName(&ast.Name{
	// 				Value: "Limit",
	// 			}),
	// 			Directives: []*ast.Directive{},
	// 			Type: ast.NewNamed(&ast.Named{
	// 				Name: ast.NewName(&ast.Name{
	// 					Value: "Int",
	// 				}),
	// 			}),
	// 		}),
	// 		ast.NewInputValueDefinition(&ast.InputValueDefinition{
	// 			Name: ast.NewName(&ast.Name{
	// 				Value: "Offset",
	// 			}),
	// 			Directives: []*ast.Directive{},
	// 			Type: ast.NewNamed(&ast.Named{
	// 				Name: ast.NewName(&ast.Name{
	// 					Value: "Int",
	// 				}),
	// 			}),
	// 		}),
	// 	},
	// })
	// a.file.Definitions = append(a.file.Definitions, searchInput)
}
