package fs

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/heyvito/go-giter8/lexer"
	"github.com/heyvito/go-giter8/sb"
)

type Node struct {
	Name lexer.AST
}

type TreeItem struct {
	Source string
	IsDir  bool
	Nodes  []Node
}

type state int

const (
	stateLiteral state = iota + 1
	stateName
	stateFormat
)

func prepareNodeName(rawName string) lexer.AST {
	ast := lexer.AST{}
	s := stateLiteral
	tmp := sb.New()
	format := sb.New()
	chars := []rune(rawName)
	for i, chr := range chars {
		var lastRune = ' '
		if i > 0 {
			lastRune = chars[i-1]
		}

		switch s {
		case stateLiteral:
			if chr == '$' && lastRune != '\\' {
				if tmp.Len() > 0 {
					ast = append(ast, lexer.Literal(tmp.String()))
					tmp.Reset()
				}

				s = stateName
				continue
			}
			tmp.WriteRune(chr)
		case stateName:
			if chr == '$' {
				if tmp.Len() == 0 {
					tmp.WriteRune(chr)
				} else {
					ast = append(ast, lexer.Template{
						Name:    tmp.String(),
						Options: nil,
					})
					tmp.Reset()
				}
				s = stateLiteral
				continue
			}
			if chr == '_' && lastRune == '_' {
				tmp.DeleteLast()
				s = stateFormat
				continue
			}
			tmp.WriteRune(chr)

		case stateFormat:
			if chr == '$' {
				ast = append(ast, lexer.Template{
					Name:    tmp.String(),
					Options: map[string]string{"format": format.String()},
				})
				tmp.Reset()
				format.Reset()
				s = stateLiteral
			}
			format.WriteRune(chr)
		}
	}
	if s == stateLiteral {
		if tmp.Len() > 0 {
			ast = append(ast, lexer.Literal(tmp.String()))
		}
	} else {
		ast = append(ast, lexer.Literal(fmt.Sprintf("$%s", tmp.String())))
		if format.Len() > 0 {
			ast = append(ast, lexer.Literal(fmt.Sprintf("__%s", format.String())))
		}
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
