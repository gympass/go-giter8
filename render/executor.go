package render

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unsafe"

	"github.com/gympass/go-giter8/lexer"
	"github.com/gympass/go-giter8/props"
)

var wordOnlyRegexp = regexp.MustCompile(`[^a-zA-Z0-9_]`)
var wordSpaceRegexp = regexp.MustCompile(`[^a-zA-Z0-9]`)
var snakeCaseRegexp = regexp.MustCompile(`[\s.]`)
var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

func uppercase(val string) string {
	return strings.ToUpper(val)
}
func lowercase(val string) string {
	return strings.ToLower(val)
}
func capitalize(val string) string {
	switch len(val) {
	case 0:
		return ""
	case 1:
		return strings.ToUpper(val)
	default:
		return strings.ToUpper(val[0:1]) + val[1:]
	}
}
func decapitalize(val string) string {
	switch len(val) {
	case 0:
		return ""
	case 1:
		return strings.ToLower(val)
	default:
		return strings.ToLower(val[0:1]) + val[2:]
	}
}
func startCase(val string) string {
	vars := strings.Split(val, " ")
	newVars := make([]string, len(vars))
	for i, w := range vars {
		newVars[i] = capitalize(w)
	}
	return strings.Join(newVars, " ")
}
func wordOnly(val string) string {
	return wordOnlyRegexp.ReplaceAllString(val, "")
}
func wordSpace(val string) string {
	return wordSpaceRegexp.ReplaceAllString(val, " ")
}
func upperCamel(val string) string {
	return wordOnly(startCase(val))
}
func lowerCamel(val string) string {
	return decapitalize(wordOnly(startCase(val)))
}
func hyphenate(val string) string {
	return strings.ReplaceAll(val, " ", "-")
}
func normalize(val string) string {
	return lowercase(hyphenate(val))
}
func snakeCase(val string) string {
	return snakeCaseRegexp.ReplaceAllString(val, "_")
}
func packageNaming(val string) string {
	return strings.ReplaceAll(val, " ", ".")
}
func packageDir(val string) string {
	return strings.ReplaceAll(val, ".", "/")
}
func generateRandom(val string) string {
	const n = 40
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return val + *(*string)(unsafe.Pointer(&b))
}

type Helper func(string) string

var helpers = map[string]Helper{
	"upper":           uppercase,
	"uppercase":       uppercase,
	"lower":           lowercase,
	"lowercase":       lowercase,
	"cap":             capitalize,
	"capitalize":      capitalize,
	"decap":           decapitalize,
	"decapitalize":    decapitalize,
	"start":           startCase,
	"start-case":      startCase,
	"word":            wordOnly,
	"word-only":       wordOnly,
	"space":           wordSpace,
	"word-space":      wordSpace,
	"Camel":           upperCamel,
	"upper-camel":     upperCamel,
	"camel":           lowerCamel,
	"lower-camel":     lowerCamel,
	"hyphen":          hyphenate,
	"hyphenate":       hyphenate,
	"norm":            normalize,
	"normalize":       normalize,
	"snake":           snakeCase,
	"snake-case":      snakeCase,
	"package":         packageNaming,
	"package-naming":  packageNaming,
	"packaged":        packageDir,
	"package-dir":     packageDir,
	"random":          generateRandom,
	"generate-random": generateRandom,
}

func extractFormatOptions(template *lexer.Template) []string {
	if template.Options == nil {
		return nil
	}
	v, ok := template.Options["format"]
	if !ok {
		return nil
	}

	allForms := strings.Split(v, ",")
	forms := make([]string, 0, len(allForms))
	for _, v := range allForms {
		trimmed := strings.TrimSpace(v)
		if len(trimmed) == 0 {
			continue
		}
		forms = append(forms, trimmed)
	}
	return forms
}

type Executor struct {
	props props.Pairs
}

func (e *Executor) runMethods(t *lexer.Template) (string, error) {
	val, ok := e.props.Fetch(t.Name)
	if !ok {
		return "", fmt.Errorf("property `%s' is not defined", t.Name)
	}
	opts := extractFormatOptions(t)
	for _, n := range opts {
		if fn, ok := helpers[n]; ok {
			val = fn(val)
		} else {
			return "", fmt.Errorf("formatter `%s' does not exist", n)
		}
	}
	return val, nil
}

func (e *Executor) evaluateConditionalExpression(expr, helper string) (bool, error) {
	v, ok := e.props.FetchPair(expr)
	if !ok {
		return false, nil
	}
	switch true {
	case strings.EqualFold(helper, "truthy"):
		return v.Truthy(), nil
	case strings.EqualFold(helper, "present"):
		return len(strings.TrimSpace(v.V)) != 0, nil
	}
	panic("BUG: helper allowed by lexer, but not implemented by renderer")
}

func (e *Executor) evaluateConditional(c *lexer.Conditional, r *strings.Builder) error {
	ok, err := e.evaluateConditionalExpression(c.Property, c.Helper)
	if err != nil {
		return err
	} else if ok {
		return e.execTree(c.Then, r)
	}

	for _, c := range c.ElseIf {
		return e.evaluateConditional(c, r)
	}

	if c.Else != nil {
		return e.execTree(c.Else, r)
	}

	return nil
}

func (e *Executor) execTree(tree lexer.AST, r *strings.Builder) error {
	for _, elem := range tree {
		switch v := elem.(type) {
		case *lexer.Literal:
			r.WriteString(v.String)
		case *lexer.Template:
			val, err := e.runMethods(v)
			if err != nil {
				return err
			}
			r.WriteString(val)
		case *lexer.Conditional:
			if err := e.evaluateConditional(v, r); err != nil {
				return err
			}
		}
	}
	return nil
}

// Exec takes a given AST and renders using props passed to the current
// Executor. Either returns a rendered string, or an error.
func (e *Executor) Exec(tree lexer.AST) (string, error) {
	var result strings.Builder
	if err := e.execTree(tree, &result); err != nil {
		return "", err
	}
	return result.String(), nil
}

// NewExecutor returns a new Executor using provided props.Pairs
func NewExecutor(props props.Pairs) *Executor {
	return &Executor{
		props: props,
	}
}
