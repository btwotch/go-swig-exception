package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"os"
	"strings"
)

type Method struct {
	name       string
	parameters []string
	returns    []string
}

func commaSeparatedList(l []string, elemPrefix string, extra ...string) string {
	var c string

	if len(l) > 0 {
		if elemPrefix == "" {
			c = l[0]
		} else {
			c = elemPrefix + "0"
		}

		for i, e := range l[1:] {
			if elemPrefix == "" {
				c += fmt.Sprintf(", %s", e)
			} else {
				c += fmt.Sprintf(", %s%d", elemPrefix, i+1)
			}
		}
	}

	if len(extra) == 0 {
		return c
	}
	if len(l) > 0 {
		c += ", "
	}

	c += strings.Join(extra, ", ")

	return c
}

func (m Method) returnList() string {
	return commaSeparatedList(m.returns, "")
}

func (m Method) parameterList() string {
	pl := ""
	for index, param := range m.parameters {
		pl += fmt.Sprintf("arg%d %s, ", index, param)
	}

	return pl
}

func (m Method) unpanicCaller() string {
	c := "func (u *__unpanic_wrap_struct) "
	c += m.name
	c += "("
	c += m.parameterList()
	c += ")"
	c += " "
	c += "("
	c += m.returnList()
	c += ")"
	c += "{\n"
	c += `
	defer func() {
		if r := recover(); r != nil {
			u.err = fmt.Errorf("%+v", r)
		}
	}()

`
	c += "\t"
	if len(m.returns) > 0 {
		c += "return "
	}
	c += "__unpanic_" + m.name
	// function arguments
	c += "("
	c += commaSeparatedList(m.parameters, "arg")
	c += ")\n"

	c += "}\n"

	c += "\n\n\n"

	c += "func " + m.name
	c += "("
	c += m.parameterList()
	c += ")"
	c += " "
	c += "("
	c += commaSeparatedList(m.returns, "", "error")
	c += ")"

	c += "{\n"
	c += "\tu := __unpanic_wrap_struct{}\n"

	c += "\t"
	c += commaSeparatedList(m.returns, "ret")
	if len(m.returns) > 0 {
		c += " := "
	}
	c += "u." + m.name

	// function arguments
	c += "("
	c += commaSeparatedList(m.parameters, "arg")
	c += ")\n"

	c += "\treturn "
	c += commaSeparatedList(m.returns, "ret", "u.err")
	c += "\n}\n"

	return c
}

func (m Method) trampoline(receiver string) string {
	var c string

	c += "func (r "
	c += receiver
	c += ") __unpanic_trampoline_"
	c += m.name
	c += "("
	c += m.parameterList()
	c += "__unpanic_err *error"
	c += ")"
	c += " "
	c += "("
	c += m.returnList()
	c += ") "
	c += "{"
	c += `
        defer func() {
                if r := recover(); r != nil {
                        *__unpanic_err = fmt.Errorf("%+v", r)
                }
        }()
`
	c += "\n\t"

	if len(m.returns) > 0 {
		c += "return "
	}
	c += "r.__unpanic_" + m.name
	c += "("
	c += commaSeparatedList(m.parameters, "arg")
	c += ")\n"

	c += "}\n\n"

	c += "func (r "
	c += receiver
	c += ") "
	c += m.name
	c += "("
	c += m.parameterList()
	c += ")"
	c += " "
	c += "("
	c += commaSeparatedList(m.returns, "", "error")
	c += ")"
	c += "{\n"

	c += "\tvar err error\n"

	c += "\t"

	c += commaSeparatedList(m.returns, "ret")
	if len(m.returns) > 0 {
		c += " := "
	}
	c += "r.__unpanic_trampoline_" + m.name
	c += "("
	c += commaSeparatedList(m.parameters, "arg", "&err")
	c += ")\n"

	c += "\treturn "
	c += commaSeparatedList(m.returns, "ret", "err")
	c += "\n}\n"
	return c
}

type PanicWrapper struct {
	// map from receiver to method
	methods map[string][]Method
}

func (pw *PanicWrapper) walkDecls(node ast.Node) {

	switch n := node.(type) {
	case *ast.GenDecl:
		for _, spec := range n.Specs {
			switch s := spec.(type) {
			case *ast.TypeSpec:
				i, ok := s.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}
				for _, field := range i.Methods.List {
					f, ok := field.Type.(*ast.FuncType)
					if !ok {
						continue
					}
					if strings.HasPrefix(field.Names[0].Name, "Swig") {
						continue
					}

					errorFuncResult := &ast.Field{}
					errorFuncResultIdent := ast.NewIdent("error")
					errorFuncResult.Names = append(errorFuncResult.Names, ast.NewIdent("__unpanic_err"))
					errorFuncResult.Type = errorFuncResultIdent
					f.Results.List = append(f.Results.List, errorFuncResult)

				}
			}
		}
	case *ast.FuncDecl:
		method := Method{}
		method.name = n.Name.Name

		if strings.HasPrefix(n.Name.Name, "Swig") {
			return
		}
		if n.Type.Results != nil {
			for _, y := range n.Type.Results.List {
				if i, ok := y.Type.(*ast.Ident); ok {
					method.returns = append(method.returns, i.Name)
				}
			}
		}
		if n.Type.Params != nil {
			for _, y := range n.Type.Params.List {
				if i, ok := y.Type.(*ast.Ident); ok {
					method.parameters = append(method.parameters, i.Name)
				}
			}
		}
		if n.Recv != nil {
			if i, ok := n.Recv.List[0].Type.(*ast.Ident); ok {
				pw.methods[i.Name] = append(pw.methods[i.Name], method)
			} else if s, ok := n.Recv.List[0].Type.(*ast.StarExpr); ok {
				if i, ok := s.X.(*ast.Ident); ok {
					pw.methods["*"+i.Name] = append(pw.methods["*"+i.Name], method)
				}
			}
		} else {
			if n.Name.Name == "main" && n.Recv == nil {
				return
			}
			pw.methods[""] = append(pw.methods[""], method)
		}
		n.Name = ast.NewIdent("__unpanic_" + n.Name.Name)
	}
}

func (pw *PanicWrapper) walk(node ast.Node) {
	switch n := node.(type) {
	case *ast.File:
		for _, f := range n.Decls {
			pw.walkDecls(f)
		}
	}
}

func newPanicWrapper() *PanicWrapper {
	pw := PanicWrapper{}
	pw.methods = make(map[string][]Method)

	return &pw
}

func main() {
	var buf bytes.Buffer

	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <path to go file>\n", os.Args[0])
		return
	}
	filename := os.Args[1]

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		panic("Could not open " + filename)
	}

	astutil.AddNamedImport(fset, file, "", "fmt")

	pw := newPanicWrapper()
	pw.walk(file)

	buf.Reset()
	printer.Fprint(&buf, fset, file)
	fmt.Println(buf.String())

	fmt.Printf(`
type __unpanic_wrap_struct struct {
	err error
}

`)
	for r, ms := range pw.methods {
		for _, m := range ms {
			if r == "" {
				fmt.Printf("%s\n", m.unpanicCaller())
			} else {
				fmt.Printf("%s\n", m.trampoline(r))
			}
		}
	}

}
