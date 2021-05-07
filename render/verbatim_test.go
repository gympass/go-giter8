package render

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Gympass/go-giter8/fs"
)

func TestVerbatim(t *testing.T) {
	rawPatterns := []string{"*.css", "*.html", "foobar.xml", "test/foo/bar.c"}
	expectations := map[string]bool{
		"hello/foo/bar.c":            false,
		"file.css":                   true,
		"foo/bar.css":                true,
		"a/longer/path/to/file.html": true,
		"something.go":               false,
		"foobar.xml":                 true,
		"other.xml":                  false,
		"/something/test/foo/bar.c":  true,
	}
	var patterns []*regexp.Regexp
	for _, r := range rawPatterns {
		patterns = append(patterns, fs.CreateSGlob(r))
	}
	for k, v := range expectations {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, v, isVerbatim(k, patterns))
		})
	}
}
