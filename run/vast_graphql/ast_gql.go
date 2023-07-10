package vast_graphql

import (
	"fmt"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/printer"

	"github.com/vision-cli/common/cases"
	"github.com/vision-cli/common/transpiler/model"
)

type AstGql struct {
	name string
	file *ast.Document
}

func NewGqlAst(name string) *AstGql {
	return &AstGql{
		name: name,
		file: ast.NewDocument(&ast.Document{
			Definitions: []ast.Node{},
		}),
	}
}

func (a *AstGql) AddType(name string, fields []model.Field, isFilterStructRequired bool) {
	newObjDef := ast.NewObjectDefinition(&ast.ObjectDefinition{
		Name: ast.NewName(&ast.Name{
			Value: cases.Pascal(a.name + name),
		}),
		Directives: []*ast.Directive{},
		Interfaces: []*ast.Named{},
		Fields:     fieldsToAstFieldDefns(fields),
	})
	a.file.Definitions = append(a.file.Definitions, newObjDef)
	a.file.Definitions = append(a.file.Definitions, addMultiObject(cases.Pascal(a.name+name)))
	a.ExtendQuery(name, isFilterStructRequired)
	a.ExtendMutation(name, fields)
}

func (a *AstGql) ExtendQuery(name string, isFilterStructRequired bool) {
	newObjDef := ast.NewTypeExtensionDefinition(&ast.TypeExtensionDefinition{
		Definition: ast.NewObjectDefinition(&ast.ObjectDefinition{
			Name: ast.NewName(&ast.Name{
				Value: "Query",
			}),
			Directives: []*ast.Directive{},
			Interfaces: []*ast.Named{},
			Fields: []*ast.FieldDefinition{
				getQueryAst(cases.Camel(a.name+"Get"+name), cases.Pascal(a.name+name)),
				listQueryAst(cases.Camel(a.name+"List"+name+"s"), "Multi"+cases.Pascal(a.name+name), name, isFilterStructRequired), // TODO: bring back when resolver is ready
			},
		}),
	})
	a.file.Definitions = append(a.file.Definitions, newObjDef)
}

func (a *AstGql) ExtendMutation(name string, fields []model.Field) {
	newObjDef := ast.NewTypeExtensionDefinition(&ast.TypeExtensionDefinition{
		Definition: ast.NewObjectDefinition(&ast.ObjectDefinition{
			Name: ast.NewName(&ast.Name{
				Value: "Mutation",
			}),
			Directives: []*ast.Directive{},
			Interfaces: []*ast.Named{},
			Fields: []*ast.FieldDefinition{
				// Create method
				ast.NewFieldDefinition(&ast.FieldDefinition{
					Name: ast.NewName(&ast.Name{
						Value: cases.Camel(a.name + "Create" + name),
					}),
					Arguments:  fieldsToAstInputValues(fields[1:]), // first element is id
					Directives: []*ast.Directive{},
					Type: ast.NewNamed(&ast.Named{
						Name: ast.NewName(&ast.Name{
							Value: "ID!",
						}),
					}),
				}),
				// Update method
				ast.NewFieldDefinition(&ast.FieldDefinition{
					Name: ast.NewName(&ast.Name{
						Value: cases.Camel(a.name + "Update" + name),
					}),
					Arguments:  fieldsToAstInputValues(fields),
					Directives: []*ast.Directive{},
					Type: ast.NewNamed(&ast.Named{
						Name: ast.NewName(&ast.Name{
							Value: "String!",
						}),
					}),
				}),
				// Delete method
				ast.NewFieldDefinition(&ast.FieldDefinition{
					Name: ast.NewName(&ast.Name{
						Value: cases.Camel(a.name + "Delete" + name),
					}),
					Arguments: []*ast.InputValueDefinition{
						ast.NewInputValueDefinition(&ast.InputValueDefinition{
							Name: ast.NewName(&ast.Name{
								Value: "id",
							}),
							Directives: []*ast.Directive{},
							Type: ast.NewNamed(&ast.Named{
								Name: ast.NewName(&ast.Name{
									Value: "ID!",
								}),
							}),
						}),
					},
					Directives: []*ast.Directive{},
					Type: ast.NewNamed(&ast.Named{
						Name: ast.NewName(&ast.Name{
							Value: "String!",
						}),
					}),
				}),
			},
		}),
	})
	a.file.Definitions = append(a.file.Definitions, newObjDef)
}

func (a *AstGql) String() string {
	return fmt.Sprintf("%v", printer.Print(a.file))
}

func getQueryAst(funcName, returnType string) *ast.FieldDefinition {
	return ast.NewFieldDefinition(&ast.FieldDefinition{
		Name: ast.NewName(&ast.Name{
			Value: funcName,
		}),
		Arguments: []*ast.InputValueDefinition{
			ast.NewInputValueDefinition(&ast.InputValueDefinition{
				Name: ast.NewName(&ast.Name{
					Value: "id",
				}),
				Directives: []*ast.Directive{},
				Type: ast.NewNamed(&ast.Named{
					Name: ast.NewName(&ast.Name{
						Value: "ID!",
					}),
				}),
			}),
		},
		Directives: []*ast.Directive{},
		Type: ast.NewNamed(&ast.Named{
			Name: ast.NewName(&ast.Name{
				Value: returnType,
			}),
		}),
	})
}

func listQueryAst(funcName, returnType, name string, isFilterStructRequired bool) *ast.FieldDefinition {
	def := ast.NewFieldDefinition(&ast.FieldDefinition{
		Name: ast.NewName(&ast.Name{
			Value: funcName,
		}),
		Arguments: []*ast.InputValueDefinition{
			ast.NewInputValueDefinition(&ast.InputValueDefinition{
				Name: ast.NewName(&ast.Name{
					Value: "pagination",
				}),
				Directives: []*ast.Directive{},
				Type: ast.NewNamed(&ast.Named{
					Name: ast.NewName(&ast.Name{
						Value: "PaginationInput",
					}),
				}),
			}),
			ast.NewInputValueDefinition(&ast.InputValueDefinition{
				Name: ast.NewName(&ast.Name{
					Value: "searchBy",
				}),
				Directives: []*ast.Directive{},
				Type: ast.NewNamed(&ast.Named{
					Name: ast.NewName(&ast.Name{
						Value: "String",
					}),
				}),
			}),
		},
		Directives: []*ast.Directive{},
		Type: ast.NewNamed(&ast.Named{
			Name: ast.NewName(&ast.Name{
				Value: returnType + "!",
			}),
		}),
	})
	if isFilterStructRequired {
		def.Arguments = append(def.Arguments, ast.NewInputValueDefinition(&ast.InputValueDefinition{
			Name: ast.NewName(&ast.Name{
				Value: "filterBy",
			}),
			Directives: []*ast.Directive{},
			Type: ast.NewNamed(&ast.Named{
				Name: ast.NewName(&ast.Name{
					Value: "Filter" + name,
				}),
			}),
		}))
	}
	return def
}

func addMultiObject(name string) *ast.ObjectDefinition {
	return ast.NewObjectDefinition(&ast.ObjectDefinition{
		Name: ast.NewName(&ast.Name{
			Value: "Multi" + name,
		}),
		Directives: []*ast.Directive{},
		Interfaces: []*ast.Named{},
		Fields: []*ast.FieldDefinition{
			ast.NewFieldDefinition(&ast.FieldDefinition{
				Name: ast.NewName(&ast.Name{
					Value: cases.Camel(name) + "s",
				}),
				Arguments:  []*ast.InputValueDefinition{},
				Directives: []*ast.Directive{},
				Type: ast.NewNamed(&ast.Named{
					Name: ast.NewName(&ast.Name{
						Value: "[" + name + "]!",
					}),
				}),
			}),
			ast.NewFieldDefinition(&ast.FieldDefinition{
				Name: ast.NewName(&ast.Name{
					Value: "pagination",
				}),
				Arguments:  []*ast.InputValueDefinition{},
				Directives: []*ast.Directive{},
				Type: ast.NewNamed(&ast.Named{
					Name: ast.NewName(&ast.Name{
						Value: "Pagination",
					}),
				}),
			}),
		},
	})
}

func fieldsToAstFieldDefns(fields []model.Field) []*ast.FieldDefinition {
	fieldDefs := []*ast.FieldDefinition{}
	for _, f := range fields {
		fieldDefs = append(fieldDefs, ast.NewFieldDefinition(&ast.FieldDefinition{
			Name: ast.NewName(&ast.Name{
				Value: cases.Camel(f.Name),
			}),
			Arguments:  []*ast.InputValueDefinition{},
			Directives: []*ast.Directive{},
			Type: ast.NewNamed(&ast.Named{
				Name: ast.NewName(&ast.Name{
					Value: f.GqlType(),
				}),
			}),
		}))
	}
	return fieldDefs
}

func fieldsToAstInputValues(fields []model.Field) []*ast.InputValueDefinition {
	inputVals := []*ast.InputValueDefinition{}
	for _, f := range fields {
		inputVals = append(inputVals, ast.NewInputValueDefinition(&ast.InputValueDefinition{
			Name: ast.NewName(&ast.Name{
				Value: cases.Camel(f.Name),
			}),
			Directives: []*ast.Directive{},
			Type: ast.NewNamed(&ast.Named{
				Name: ast.NewName(&ast.Name{
					Value: f.GqlType(),
				}),
			}),
		}))
	}
	return inputVals
}
