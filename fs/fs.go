package fs

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/Gympass/go-giter8/lexer"
)

type Node struct {
	Name lexer.AST
}

type TreeItem struct {
	Source string
	IsDir  bool
	Nodes  []Node
}

func prepareNodeName(rawName string) lexer.AST {
	ast, err := lexer.Tokenize(rawName)
	if err != nil {
		return lexer.AST{lexer.Literal{String: rawName}}
	}
	return ast
}

// ScanTree takes a source directory and returns a slice of TreeItem
// ready to be processed by a renderer
func ScanTree(source string) ([]TreeItem, error) {
	var items []TreeItem
	sep := string(filepath.Separator)
	err := filepath.Walk(source, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == source {
			return nil
		}

		var nodes []Node
		src := strings.Split(strings.TrimPrefix(path, source+sep), sep)
		for _, x := range src {
			nodes = append(nodes, Node{Name: prepareNodeName(x)})
		}
		items = append(items, TreeItem{
			Source: path,
			IsDir:  info.IsDir(),
			Nodes:  nodes,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}
