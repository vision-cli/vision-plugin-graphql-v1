package vast_graphql

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"reflect"
)

const (
	WriteMode = 0600
)

// A convenience function for parsing a Stmt by wrapping it in a program
func ParseStmt(code string) (*[]ast.Stmt, error) {
	// wrap the code in a program so we can parse it
	prog := "package main\nfunc main() {\n" + code + "\n}"
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", prog, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	expr := file.Decls[0].(*ast.FuncDecl).Body.List
	return &expr, nil
}

// A convenience function for printing a Stmt by wrapping it in a program and printing it
func PrintStmt(stmt ast.Stmt) (string, error) {
	// wrap the code in a program so we can parse it
	prog := "package main\nfunc main() {}"
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", prog, parser.ParseComments)
	if err != nil {
		return "", err
	}
	file.Decls[0].(*ast.FuncDecl).Body.List = append(file.Decls[0].(*ast.FuncDecl).Body.List, stmt)
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, file); err != nil {
		return "", err
	}
	str := buf.String()
	return str[28 : len(str)-3], nil
}

// areStmtsEqual compares two ast.Stmt while ignoring positions
func areStmtsEqual(stmt1, stmt2 ast.Stmt) bool {
	if stmt1 == nil || stmt2 == nil {
		return stmt1 == stmt2
	}

	// Compare the statement types
	if reflect.TypeOf(stmt1) != reflect.TypeOf(stmt2) {
		return false
	}

	switch stmt1 := stmt1.(type) {
	case *ast.AssignStmt:
		if stmt2, ok := stmt2.(*ast.AssignStmt); ok {
			return areExprsEqualList(stmt1.Lhs, stmt2.Lhs) &&
				areExprsEqualList(stmt1.Rhs, stmt2.Rhs) &&
				stmt1.Tok == stmt2.Tok
		}
	case *ast.ExprStmt:
		if stmt2, ok := stmt2.(*ast.ExprStmt); ok {
			return areExprsEqual(stmt1.X, stmt2.X)
		}
	case *ast.ReturnStmt:
		if stmt2, ok := stmt2.(*ast.ReturnStmt); ok {
			return areExprsEqualList(stmt1.Results, stmt2.Results)
		}
		// Add cases for other statement types you want to compare

	// Add cases for other statement types you want to compare

	default:
		// Handle any other specific statement types you want to compare
		// or return false if you don't want to compare them
	}

	return false
}

// areAssignStmtsEqual compares two ast.AssignStmt while ignoring positions
//
//lint:ignore U1000 Ignore unused function
func areAssignStmtsEqual(stmt1, stmt2 *ast.AssignStmt) bool {
	if stmt1 == nil || stmt2 == nil {
		return stmt1 == stmt2
	}

	// Compare the number of Lhs and Rhs expressions
	if len(stmt1.Lhs) != len(stmt2.Lhs) || len(stmt1.Rhs) != len(stmt2.Rhs) {
		return false
	}

	// Compare the types of the assignment tokens
	if stmt1.Tok != stmt2.Tok {
		return false
	}

	// Compare the Lhs expressions
	for i := range stmt1.Lhs {
		if !areExprsEqual(stmt1.Lhs[i], stmt2.Lhs[i]) {
			return false
		}
	}

	// Compare the Rhs expressions
	for i := range stmt1.Rhs {
		if !areExprsEqual(stmt1.Rhs[i], stmt2.Rhs[i]) {
			return false
		}
	}

	return true
}

// areExprsEqual compares two ast.Expr while ignoring positions
func areExprsEqual(expr1, expr2 ast.Expr) bool {
	if expr1 == nil || expr2 == nil {
		return expr1 == expr2
	}

	// Compare the expression types
	if reflect.TypeOf(expr1) != reflect.TypeOf(expr2) {
		return false
	}

	switch expr1 := expr1.(type) {
	case *ast.Ident:
		if expr2, ok := expr2.(*ast.Ident); ok {
			return expr1.Name == expr2.Name
		}
	case *ast.BasicLit:
		if expr2, ok := expr2.(*ast.BasicLit); ok {
			return expr1.Value == expr2.Value
		}
	case *ast.BinaryExpr:
		if expr2, ok := expr2.(*ast.BinaryExpr); ok {
			return expr1.Op == expr2.Op &&
				areExprsEqual(expr1.X, expr2.X) &&
				areExprsEqual(expr1.Y, expr2.Y)
		}
		// Add cases for other expression types you want to compare

	// Add cases for other expression types you want to compare

	default:
		// Handle any other specific expression types you want to compare
		// or return false if you don't want to compare them
	}

	return false
}

// areStmtsEqual compares two ast.Stmt while ignoring positions
func areStmtsEqualList(list1, list2 []ast.Stmt) bool {
	if len(list1) != len(list2) {
		return false
	}

	for i := range list1 {
		if !areStmtsEqual(list1[i], list2[i]) {
			return false
		}
	}

	return true
}

// areExprsEqualList compares two lists of ast.Expr while ignoring positions
func areExprsEqualList(list1, list2 []ast.Expr) bool {
	if len(list1) != len(list2) {
		return false
	}

	for i := range list1 {
		if !areExprsEqual(list1[i], list2[i]) {
			return false
		}
	}

	return true
}
