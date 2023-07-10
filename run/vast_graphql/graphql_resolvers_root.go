package vast_graphql

import (
	"go/ast"
	"io/ioutil"
	"path/filepath"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/vision-cli/common/cases"
	"github.com/vision-cli/common/transpiler/model"
)

func UpdateRootResolver(service model.Service, projectRoot string, moduleName string) {
	path := filepath.Join(projectRoot, "services", "default", "graphql", "resolvers", "root.go")
	astRootResolver := NewAstFromFile(path)

	astRootResolver = addFieldToRootStruct(astRootResolver, service.Name, moduleName)
	astRootResolver = addResolverToReturn(astRootResolver, service.Name, moduleName)

	err := ioutil.WriteFile(path, []byte(astRootResolver.String()), WriteMode)
	if err != nil {
		panic(err)
	}
}

func addFieldToRootStruct(a *AstGo, serviceName, moduleName string) *AstGo {
	for _, decl := range a.file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if structType, ok := typeSpec.Type.(*ast.StructType); ok && typeSpec.Name.Name == "Root" {
						newField := &ast.Field{
							Names: []*ast.Ident{
								ast.NewIdent(cases.Pascal(moduleName) + cases.Pascal(serviceName) + "Resolver"),
							},
							Type: ast.NewIdent(""),
						}

						structType.Fields.List = append(structType.Fields.List, newField)
						break
					}
				}
			}
		}
	}
	return a
}

func addResolverToReturn(a *AstGo, serviceName, moduleName string) *AstGo {
	astutil.Apply(a.file, nil, func(c *astutil.Cursor) bool {
		if funcDecl, ok := c.Node().(*ast.FuncDecl); ok && funcDecl.Name.Name == "NewRoot" {
			funcDecl.Body.List[0].(*ast.ReturnStmt).Results[0].(*ast.UnaryExpr).X.(*ast.CompositeLit).Elts = append(
				funcDecl.Body.List[0].(*ast.ReturnStmt).Results[0].(*ast.UnaryExpr).X.(*ast.CompositeLit).Elts,
				&ast.KeyValueExpr{
					Key:   ast.NewIdent(cases.Pascal(moduleName) + cases.Pascal(serviceName) + "Resolver"),
					Value: ast.NewIdent("New" + cases.Pascal(moduleName) + cases.Pascal(serviceName) + "Resolver()"),
				},
			)
			return false // Stop the traversal
		}
		return true // Continue traversing
	})
	return a
}
