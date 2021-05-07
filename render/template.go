package render

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Gympass/go-giter8/fs"
	"github.com/Gympass/go-giter8/lexer"
	"github.com/Gympass/go-giter8/props"
)

func isText(s []byte) bool {
	const max = 1024 // at least utf8.UTFMax
	if len(s) > max {
		s = s[0:max]
	}
	for i, c := range string(s) {
		if i+utf8.UTFMax > len(s) {
			// last char may be incomplete - ignore
			break
		}
		if c == 0xFFFD || c < ' ' && c != '\n' && c != '\t' && c != '\f' {
			// decoding error or control character - not a text file
			return false
		}
	}
	return true
}

func isTextFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	var buf [1024]byte
	n, err := f.Read(buf[0:])
	if err != nil {
		return false
	}

	return isText(buf[0:n])
}

func copyFile(src, dst string) error {
	sourceStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	if _, err = os.Stat(dst); err == nil {
		return fmt.Errorf("%s: destination file already exists", dst)
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if err = destination.Chmod(sourceStat.Mode()); err != nil {
		return err
	}

	buf := make([]byte, 4096)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err = destination.Write(buf[:n]); err != nil {
			return err
		}
	}

	return nil
}

func renderAndJoin(exec *Executor, nodes []fs.Node) (string, error) {
	var items []string
	for _, n := range nodes {
		if r, err := exec.Exec(n.Name); err != nil {
			return "", err
		} else {
			if r == "" {
				// If a single item yields an empty string, we can safely
				// invalidate all the path and do not work on this fs node
				return "", nil
			}
			items = append(items, r)
		}
	}
	return filepath.Join(items...), nil
}

func isVerbatim(source string, patterns []*regexp.Regexp) bool {
	for _, p := range patterns {
		if p.MatchString(source) {
			return true
		}
	}
	return false
}

// TemplateDirectory renders a given source template using props as variables
// into a given destination. Destination must not exist.
func TemplateDirectory(props props.Pairs, source, destination string) error {
	items, err := fs.ScanTree(source)
	if err != nil {
		return err
	}

	// Stat destination...
	_, err = os.Stat(destination)
	if err == nil {
		return fmt.Errorf("destination %s already exists", destination)
	} else if !os.IsNotExist(err) {
		return err
	}

	if err = os.MkdirAll(destination, os.ModePerm); err != nil {
		return err
	}

	exec := NewExecutor(props)
	verb, verbOK := props.Fetch("verbatim")
	var verbs []*regexp.Regexp
	for _, v := range strings.Split(verb, " ") {
		v = strings.TrimSpace(v)
		if len(v) > 0 {
			reg := fs.CreateSGlob(v)
			verbs = append(verbs, reg)
		}
	}

	for _, item := range items {
		path, err := renderAndJoin(exec, item.Nodes)
		if err != nil {
			return err
		}
		if path == "" {
			continue
		}
		path = filepath.Join(destination, path)

		if item.IsDir {
			if err = os.MkdirAll(path, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		if (verbOK && isVerbatim(item.Source, verbs)) || !isTextFile(item.Source) {
			// Just... copy it?
			if err = copyFile(item.Source, path); err != nil {
				return err
			}
			continue
		}
		fileStat, err := os.Stat(item.Source)
		if err != nil {
			return err
		}
		fileContents, err := os.ReadFile(item.Source)
		if err != nil {
			return err
		}
		ast, err := lexer.Tokenize(string(fileContents))
		if err != nil {
			return fmt.Errorf("error parsing %s: %s", item.Source, err)
		}

		contents, err := exec.Exec(ast)
		if err != nil {
			return fmt.Errorf("error rendering %s: %s", item.Source, err)
		}
		err = os.WriteFile(path, []byte(contents), fileStat.Mode())
		if err != nil {
			return err
		}
	}

	return nil
}
