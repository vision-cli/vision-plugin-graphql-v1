package vast_graphql

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/vision-cli/common/transpiler/model"
)

type AstGo struct {
	fset *token.FileSet
	file *ast.File
}

func NewAstFromString(code string) *AstGo {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", code, 0)
	if err != nil {
		panic(err)
	}
	return &AstGo{
		fset: fset,
		file: file,
	}
}

func NewAstFromFile(filename string) *AstGo {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		panic(err)
	}
	return &AstGo{
		fset: fset,
		file: file,
	}
}

func NewBlankAst(identity string, service model.Service) *AstGo {
	fset := token.NewFileSet()
	file := &ast.File{
		Name:  ast.NewIdent(identity),
		Scope: ast.NewScope(nil),
	}

	a := AstGo{
		fset: fset,
		file: file,
	}

	if service.HasTimestamp() {
		a.addImport("time")
	}
	if service.HasPersistence() {
		a.addImport("gorm.io/gorm")
	}
	// we always had Ids
	a.addImport("github.com/google/uuid")

	return &a
}

func NewCompletlyBlankAst(identity string, service model.Service) *AstGo {
	fset := token.NewFileSet()
	file := &ast.File{
		Name:  ast.NewIdent(identity),
		Scope: ast.NewScope(nil),
	}

	a := AstGo{
		fset: fset,
		file: file,
	}
	return &a
}

func (a *AstGo) addImport(path string) {
	astutil.AddImport(a.fset, a.file, path)
}

func (a *AstGo) AddImportAs(path, as string) {
	// Create the import statement
	importSpec := &ast.ImportSpec{
		Name: ast.NewIdent(as),
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: `"` + path + `"`,
		},
	}

	// Add the import statement to the AST file
	importDecl := &ast.GenDecl{
		Tok:    token.IMPORT,
		Specs:  []ast.Spec{importSpec},
		Lparen: 1, // Set to 1 to indicate no parenthesis in import declaration
		Rparen: 1,
	}
	a.file.Decls = append(a.file.Decls, importDecl)
}

func (a *AstGo) RemoveFunc(name string) {
	astutil.Apply(a.file, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		switch n.(type) {
		case *ast.FuncDecl:
			f, ok := n.(*ast.FuncDecl)
			if ok {
				if f.Name.String() == name {
					c.Delete()
				}
			}
		}
		return true
	})
}

func (a *AstGo) String() string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, a.fset, a.file); err != nil {
		panic(err)
	}

	return buf.String()
}
