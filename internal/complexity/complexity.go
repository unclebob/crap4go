package complexity

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type Function struct {
	Name       string
	Package    string
	StartLine  int
	EndLine    int
	Complexity int
}

func ExtractFunctions(path string) ([]Function, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}
	var functions []Function
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		functions = append(functions, Function{
			Name:       functionName(fn),
			Package:    file.Name.Name,
			StartLine:  fset.Position(fn.Pos()).Line,
			EndLine:    fset.Position(fn.End()).Line,
			Complexity: CyclomaticComplexity(fn),
		})
	}
	return functions, nil
}

func CyclomaticComplexity(fn *ast.FuncDecl) int {
	complexity := 1
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
			complexity++
		case *ast.CaseClause:
			complexity++
		case *ast.CommClause:
			complexity++
		case *ast.BinaryExpr:
			if n.Op == token.LAND || n.Op == token.LOR {
				complexity++
			}
		}
		return true
	})
	return complexity
}

func functionName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return fn.Name.Name
	}
	return receiverName(fn.Recv.List[0].Type) + "." + fn.Name.Name
}

func receiverName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return receiverName(t.X)
	case *ast.IndexExpr:
		return receiverName(t.X)
	case *ast.IndexListExpr:
		return receiverName(t.X)
	case *ast.SelectorExpr:
		return receiverName(t.X) + "." + t.Sel.Name
	default:
		return strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(exprString(expr)), "*"), "&")
	}
}

func exprString(expr ast.Expr) string {
	var b strings.Builder
	ast.Fprint(&b, token.NewFileSet(), expr, nil)
	return b.String()
}
