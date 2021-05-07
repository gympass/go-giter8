package fs

import (
	"regexp"
	"strings"
)

// sglob is a small glob handler
var (
	beginRegexp     = regexp.MustCompile(`([^/+])/.*\*\.`)
	dotEscapeRegexp = regexp.MustCompile(`\.`)
)

const magicStar = "#$~"

type repl struct {
	reg     *regexp.Regexp
	replace string
}

var repls = []repl{
	{regexp.MustCompile(`/\*\*/`), `(/|/.+/)`},
	{regexp.MustCompile(`\*\*/`), `(|.` + magicStar + `/)`},
	{regexp.MustCompile(`/\*\*`), `(|/.` + magicStar + `)`},
	{regexp.MustCompile(`\\\*`), `\` + magicStar},
	{regexp.MustCompile(`\*`), `([^/]*)`},
}

func CreateSGlob(line string) *regexp.Regexp {
	line = strings.TrimRight(line, "\r")
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	if beginRegexp.MatchString(line) && line[0] != '/' {
		line = "/" + line
	}
	line = dotEscapeRegexp.ReplaceAllString(line, `\.`)
	if strings.HasPrefix(line, "/**/") {
		line = line[1:]
	}
	for _, r := range repls {
		line = r.reg.ReplaceAllString(line, r.replace)
	}
	line = strings.Replace(line, "?", `\?`, -1)
	line = strings.Replace(line, magicStar, "*", -1)

	var expr = ""
	if strings.HasSuffix(line, "/") {
		expr = line + "(|.*)$"
	} else {
		expr = line + "(|/.*)$"
	}
	if strings.HasPrefix(expr, "/") {
		expr = "^(|/)" + expr[1:]
	} else {
		expr = "^(|.*/)" + expr
	}
	pattern, _ := regexp.Compile(expr)

	return pattern
}
