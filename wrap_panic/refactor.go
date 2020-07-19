package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
)

func walkCallExpr(node ast.Node) {
	switch n := node.(type) {
	case *ast.Ident:
		fmt.Printf("default: %s\n", n)
		if n.String() == "panic" {
			fmt.Printf("P\n")
		}
	default:
		fmt.Printf("default: %T\n", n)
	}
}

func walkExpr(node ast.Node) {
	switch n := node.(type) {

	case *ast.CallExpr:
		walkCallExpr(n.Fun)
	}
}

func walkStmt(node ast.Node) {
	switch n := node.(type) {
	case *ast.ExprStmt:
		walkExpr(n.X)
	}
}

func stmtIsPanic(node ast.Node) bool {
	if _, ok := node.(*ast.ExprStmt); !ok {
		return false
	}
	node = node.(*ast.ExprStmt).X
	if _, ok := node.(*ast.CallExpr); !ok {
		return false
	}
	node = node.(*ast.CallExpr).Fun
	if _, ok := node.(*ast.Ident); !ok {
		return false
	}

	if node.(*ast.Ident).String() == "panic" {
		fmt.Printf("P\n")
		return true
	}
	return false
}

// returns true if has panic
func walkFunc(node ast.Node) bool {
	switch n := node.(type) {

	case *ast.BlockStmt:
		for i, x := range n.List {
			if walkFunc(x) {
				return true
			}
			//walkStmt(x)
			if stmtIsPanic(x) {
				/*
					retstmt := new(ast.ReturnStmt)
					retstmt.Results = append(retstmt.Results, ast.NewIdent("nil"))
					retstmt.Results = append(retstmt.Results, ast.NewIdent("fmt.Errorf(\"this would have been a panic\")"))
					n.List[i] = retstmt
				*/
				fmt.Println(fmt.Sprintf("panic found %d", i))
				return true
			}
		}
	case *ast.IfStmt:
		if walkFunc(n.Body) {
			return true
		}
		//default:
		//	fmt.Printf("default: %T\n", n)
	}

	return false
}

func walkDecls(node ast.Node) {
	switch n := node.(type) {
	case *ast.FuncDecl:
		fmt.Printf(">>>>>>>> func decl: %s\n", n.Name)
		// only change methods that don't have a receiver
		if n.Recv != nil {
			fmt.Printf("\ton %+v\n", n.Recv.List[0])
			f := &ast.Field{}
			f.Type = ast.NewIdent(fmt.Sprintf("__unpanic_%s", n.Recv.List[0].Type))
			//f.Type = ast.NewIdent("__unpanic_blaasdf")
			f.Names = append(f.Names, ast.NewIdent(n.Recv.List[0].Names[0].Name))
			n.Recv.List[0] = f
			return
		}
		if n.Name.Name == "main" {
			return
		}
		n.Name = ast.NewIdent("__unpanic_" + n.Name.Name)
/*
		if walkFunc(n.Body) {
			fmt.Printf("panic: %+v\n", n)
			if n.Type.Results != nil {
				//	f := &ast.Field{}
				//	f.Type = ast.NewIdent("error")
				//	n.Type.Results.List = append(n.Type.Results.List, f)
				fmt.Println("returns found")
			}
		}
		if n.Type.Results != nil {
			for _, x := range n.Type.Results.List {
				fmt.Printf("type results: %+v\n", x)
			}
			//fmt.Printf("func: %+v\n", *n.Type.Results.List)
		}
*/
	}
}

func walk(node ast.Node) {
	switch n := node.(type) {
	case *ast.File:
		for _, f := range n.Decls {
			walkDecls(f)
		}
	}
}

func main() {
	var buf bytes.Buffer
	filename := "test/exception.go"

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		panic("Could not open " + filename)
	}

	//ast.Print(fset, file)
	walk(file)

	fmt.Println("--------------------------------")
	buf.Reset()
	printer.Fprint(&buf, fset, file)
	fmt.Println(buf.String())
}
