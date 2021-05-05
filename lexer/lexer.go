package lexer

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/Gympass/go-giter8/sb"
)

type Kind int

const (
	KindInvalid Kind = iota
	KindLiteral
	KindTemplate
	KindConditional
)

const DEBUG = false

func debug(format string, a ...interface{}) {
	if DEBUG {
		fmt.Printf("%s\n", fmt.Sprintf(format, a...))
	}
}

type Node interface {
	Kind() Kind
	Parent() Node
}

type Literal struct {
	String     string
	nodeParent Node
}

func (l Literal) Kind() Kind {
	return KindLiteral
}

func (l Literal) Parent() Node {
	return l.nodeParent
}

type Template struct {
	Name       string
	Options    map[string]string
	nodeParent Node
}

func (t Template) Parent() Node {
	return t.nodeParent
}

func (t Template) Kind() Kind {
	return KindTemplate
}

type Conditional struct {
	Property   string
	Helper     string
	Then       AST
	ElseIf     []*Conditional
	Else       AST
	parentNode Node
}

func (c Conditional) Kind() Kind {
	return KindConditional
}

func (c Conditional) Parent() Node {
	return c.parentNode
}

type AST []Node

type state int

const (
	stateLiteral state = iota
	stateTemplateName
	stateTemplateCombinedFormatter
	stateTemplateConditionalExpression
	stateTemplateConditionalExpressionEnd
	stateTemplateConditionalThen
	stateTemplateConditionalElseIf
	stateTemplateConditionalElse
	stateTemplateOptionName
	stateTemplateOptionValueBegin
	stateTemplateOptionValue
	stateTemplateOptionOrEnd
)

func (s state) String() string {
	switch s {
	case stateLiteral:
		return "stateLiteral"
	case stateTemplateName:
		return "stateTemplateName"
	case stateTemplateCombinedFormatter:
		return "stateTemplateCombinedFormatter"
	case stateTemplateConditionalExpression:
		return "stateTemplateConditionalExpression"
	case stateTemplateConditionalExpressionEnd:
		return "stateTemplateConditionalExpressionEnd"
	case stateTemplateConditionalThen:
		return "stateTemplateConditionalThen"
	case stateTemplateConditionalElseIf:
		return "stateTemplateConditionalElseIf"
	case stateTemplateConditionalElse:
		return "stateTemplateConditionalElse"
	case stateTemplateOptionName:
		return "stateTemplateOptionName"
	case stateTemplateOptionValueBegin:
		return "stateTemplateOptionValueBegin"
	case stateTemplateOptionValue:
		return "stateTemplateOptionValue"
	case stateTemplateOptionOrEnd:
		return "stateTemplateOptionOrEnd"
	default:
		return "WTF!"
	}
}

const (
	ESCAPE     = rune('\\')
	DELIM      = rune('$')
	NEWLINE    = rune('\n')
	SEMICOLON  = rune(';')
	EQUALS     = rune('=')
	QUOT       = rune('"')
	COMMA      = rune(',')
	SPACE      = rune(' ')
	HTAB       = rune('\t')
	LPAREN     = rune('(')
	RPAREN     = rune(')')
	DOT        = rune('.')
	UNDERSCORE = rune('_')
	TRUTHY     = "truthy"
)

func isSpace(r rune) bool {
	return r == SPACE || r == HTAB
}

type stateStack []state

func (s *stateStack) String() string {
	result := make([]string, 0, len(*s))
	for _, v := range *s {
		result = append(result, v.String())
	}
	return strings.Join(result, ", ")
}

type Tokenizer struct {
	ast AST
	tmp *sb.StringBuilder

	templateName       *sb.StringBuilder
	optionName         *sb.StringBuilder
	optionValue        *sb.StringBuilder
	templateOptions    map[string]string
	_state             state
	currentConditional *Conditional
	stateStack         stateStack

	lastFedRune rune
	idx         int
	line        int
}

// NewTokenizer prepares a new Tokenizer
func NewTokenizer() *Tokenizer {
	return &Tokenizer{
		ast:             nil,
		tmp:             sb.New(),
		templateName:    sb.New(),
		optionName:      sb.New(),
		optionValue:     sb.New(),
		templateOptions: nil,
		_state:          0,
		lastFedRune:     0,
		idx:             0,
		line:            0,
	}
}

func (t *Tokenizer) pushStack() {
	t.stateStack = append(t.stateStack, t._state)
	if DEBUG {
		debug("STS: PUSH [%s]", t.stateStack.String())
	}
}

func (t *Tokenizer) transition(s state) {
	debug("STT: Transitioning %s -> %s", t._state, s)
	t._state = s
}

func (t *Tokenizer) popStack() state {
	if len(t.stateStack) == 0 {
		panic("BUG? Attempt to pop state stack beyond limit")
	}
	t.transition(t.stateStack[len(t.stateStack)-1])
	t.stateStack = t.stateStack[0 : len(t.stateStack)-1]
	if DEBUG {
		debug("STS: POP [%s]", t.stateStack.String())
	}
	return t._state
}

func (t *Tokenizer) replaceStack(s state) {
	if len(t.stateStack) == 0 {
		panic("BUG? Attempt to replace on empty state stack")
	}
	t.stateStack[len(t.stateStack)-1] = s
	if DEBUG {
		debug("STS: REPLACE [%s]", t.stateStack.String())
	}
}

func (t *Tokenizer) currentStack() (bool, state) {
	if len(t.stateStack) == 0 {
		return false, 0
	}
	return true, t.stateStack[len(t.stateStack)-1]
}

func (t *Tokenizer) pushAST(n Node) {
	if ok, s := t.currentStack(); ok {
		if s == stateTemplateConditionalThen || s == stateTemplateConditionalElseIf {
			t.currentConditional.Then = append(t.currentConditional.Then, n)
		} else {
			t.currentConditional.Else = append(t.currentConditional.Else, n)
		}
	} else {
		t.ast = append(t.ast, n)
	}
}

func (t *Tokenizer) commitLiteral() {
	if t.tmp.Len() == 0 {
		return
	}
	t.pushAST(&Literal{String: t.tmp.String(), nodeParent: t.currentConditional})
	t.tmp.Reset()
}

func (t *Tokenizer) commitTemplate() {
	if t.templateName.Len() == 0 {
		return
	}
	t.pushAST(&Template{
		Name:       strings.TrimSpace(t.templateName.String()),
		Options:    t.templateOptions,
		nodeParent: t.currentConditional,
	})
	t.templateName.Reset()
	t.templateOptions = nil
}

func (t *Tokenizer) commitTemplateOption() {
	if t.optionName.Len() == 0 {
		return
	}

	if t.templateOptions == nil {
		t.templateOptions = map[string]string{}
	}
	t.templateOptions[strings.TrimSpace(t.optionName.String())] = t.optionValue.String()
	t.optionName.Reset()
	t.optionValue.Reset()
}

func (t *Tokenizer) prepareConditional() error {
	ok, ls := t.currentStack()
	debug("CND: Current state: %s, lastStack(%v): %s", t._state, ok, ls)
	expr := t.templateName.String()
	separatorIndex := strings.IndexRune(expr, DOT)
	if separatorIndex == -1 {
		return t.invalidConditionalExpression(expr)
	}
	var (
		prop   = expr[0:separatorIndex]
		helper = expr[separatorIndex+1:]
	)

	if !strings.EqualFold(helper, TRUTHY) {
		return t.unsupportedConditionalHelper(helper)
	}

	cond := &Conditional{
		Property:   prop,
		Helper:     helper,
		Then:       nil,
		Else:       nil,
		parentNode: t.currentConditional,
	}
	if ls == stateTemplateConditionalThen {
		t.ast = append(t.ast, cond)
	} else if ls == stateTemplateConditionalElseIf {
		t.currentConditional.ElseIf = append(t.currentConditional.ElseIf, cond)
	}
	t.currentConditional = cond
	t.templateName.Reset()
	return nil
}

func (t *Tokenizer) lastRune() rune {
	return t.lastFedRune
}

func (t Tokenizer) unexpectedToken(token rune) error {
	return UnexpectedTokenErr{idx: t.idx, line: t.line, token: string(token)}
}

func (t Tokenizer) unexpectedLineBreak() error {
	return UnexpectedLinebreakErr{line: t.line, idx: t.idx}
}

func (t Tokenizer) unexpectedKeyword(n string) error {
	return UnexpectedTokenErr{idx: t.idx, line: t.line, token: n}
}

func (t Tokenizer) invalidConditionalExpression(expr string) error {
	return InvalidConditionalExpressionErr{idx: t.idx, line: t.line, expr: expr}
}

func (t Tokenizer) unsupportedConditionalHelper(name string) error {
	return UnsupportedConditionalHelperErr{idx: t.idx, line: t.line, helper: name}
}

// Feed feeds a given rune to the parser
func (t *Tokenizer) Feed(chr rune) error {
	defer func() {
		t.idx++
		if chr == NEWLINE {
			t.line++
		}
		t.lastFedRune = chr
	}()
	if DEBUG {
		pchr := string(chr)
		if pchr == "\n" {
			pchr = "\\n"
		}
		debug("CHR: %s, STATE: %s", pchr, t._state)
	}
	switch t._state {
	case stateLiteral:
		if chr == DELIM && t.lastRune() != ESCAPE {
			t.commitLiteral()
			t.transition(stateTemplateName)
			return nil
		} else if chr == DELIM && t.lastRune() == ESCAPE {
			t.tmp.DeleteLast()
		}
		t.tmp.WriteRune(chr)

	case stateTemplateName:
		if chr == DELIM {
			if t.templateName.Len() == 0 {
				return t.unexpectedToken(DELIM)
			}
			currentName := t.templateName.String()
			if currentName == "if" {
				return t.unexpectedKeyword(currentName)
			} else if currentName == "else" {
				if ok, ss := t.currentStack(); !ok {
					return t.unexpectedKeyword(currentName)
				} else if ss == stateTemplateConditionalElseIf {
					t.popStack()
					parent := t.currentConditional.Parent()
					if parent == nil || reflect.ValueOf(parent).IsNil() {
						panic("BUG: ElseIf without parent")
					} else if parent.Kind() != KindConditional {
						panic("BUG: ElseIf without conditional parent")
					}
					t.currentConditional = parent.(*Conditional)
				}

				t.replaceStack(stateTemplateConditionalElse)
				t.transition(stateLiteral)
				t.templateName.Reset()
				return nil
			} else if currentName == "endif" {
				if ok, _ := t.currentStack(); !ok {
					return t.unexpectedKeyword(currentName)
				}
				t.popStack()
				prevCond := t.currentConditional.Parent()
				if prevCond == nil || reflect.ValueOf(prevCond).IsNil() {
					t.currentConditional = nil
				} else if prevCond.Kind() != KindConditional {
					panic("BUG? Parent is not conditional")
				}
				t.currentConditional = prevCond.(*Conditional)
				t.transition(stateLiteral)
				t.templateName.Reset()
				return nil
			}
			t.commitTemplate()
			t.transition(stateLiteral)
			return nil
		} else if chr == SPACE {
			return t.unexpectedToken(SPACE)
		} else if chr == SEMICOLON {
			t.transition(stateTemplateOptionName)
			return nil
		} else if chr == NEWLINE {
			return t.unexpectedLineBreak()
		} else if chr == LPAREN && (t.templateName.String() == "if" || t.templateName.String() == "elseif") {
			if t.templateName.String() == "if" {
				t.transition(stateTemplateConditionalThen)
			} else {
				// Transitioning to ElseIf...
				if ok, current := t.currentStack(); !ok || current == stateTemplateConditionalElse {
					// At this point we either have an elseif out of an if
					// structure, or we have an elseif after an else. Both are
					// unacceptable.
					return t.unexpectedKeyword("elseif")
				}
				t.transition(stateTemplateConditionalElseIf)
			}
			t.pushStack()
			t.transition(stateTemplateConditionalExpression)
			t.templateName.Reset()
			return nil
		}
		if t.templateName.Len() == 0 && !unicode.IsLetter(chr) {
			return t.unexpectedToken(chr)
		}
		if chr == UNDERSCORE && t.lastRune() == UNDERSCORE {
			t.templateName.DeleteLast()
			t.transition(stateTemplateCombinedFormatter)
			t.tmp.Reset()
			return nil
		}
		if !unicode.IsLetter(chr) && !unicode.IsDigit(chr) && chr != UNDERSCORE {
			return t.unexpectedToken(chr)
		}
		t.templateName.WriteRune(chr)
	case stateTemplateCombinedFormatter:
		if chr == DELIM {
			if t.tmp.Len() == 0 {
				return t.unexpectedToken(chr)
			}
			t.templateOptions = map[string]string{
				"format": t.tmp.String(),
			}
			t.commitTemplate()
			t.tmp.Reset()
			t.transition(stateLiteral)
			return nil
		}
		t.tmp.WriteRune(chr)

	case stateTemplateConditionalExpression:
		if chr == RPAREN {
			if t.templateName.Len() == 0 {
				return t.unexpectedToken(chr)
			}
			t.transition(stateTemplateConditionalExpressionEnd)
			return nil
		}
		if !unicode.IsLetter(chr) && !unicode.IsDigit(chr) && chr != DOT {
			return t.unexpectedToken(chr)
		}
		t.templateName.WriteRune(chr)

	case stateTemplateConditionalExpressionEnd:
		if chr != DELIM {
			return t.unexpectedToken(chr)
		}
		if err := t.prepareConditional(); err != nil {
			return err
		}
		t.transition(stateLiteral)

	case stateTemplateOptionName:
		if chr == DELIM {
			if t.templateName.Len() == 0 {
				return t.unexpectedToken(DELIM)
			}
			t.commitTemplate()
			t.transition(stateLiteral)
			return nil
		} else if chr == EQUALS {
			t.transition(stateTemplateOptionValueBegin)
			return nil
		}
		t.optionName.WriteRune(chr)

	case stateTemplateOptionValueBegin:
		if isSpace(chr) {
			return nil
		} else if chr == QUOT {
			t.transition(stateTemplateOptionValue)
			return nil
		}
		return t.unexpectedToken(chr)

	case stateTemplateOptionValue:
		if chr == QUOT && t.lastRune() != ESCAPE {
			t.transition(stateTemplateOptionOrEnd)
			t.commitTemplateOption()
			return nil
		} else if chr == QUOT && t.lastRune() == ESCAPE {
			t.optionValue.DeleteLast()
		}
		t.optionValue.WriteRune(chr)

	case stateTemplateOptionOrEnd:
		if isSpace(chr) {
			return nil
		} else if chr == COMMA {
			t.transition(stateTemplateOptionName)
			return nil
		} else if chr == DELIM {
			t.transition(stateLiteral)
			t.commitTemplate()
			return nil
		}
		return t.unexpectedToken(chr)
	}

	return nil
}

// Finish completes the parsing process and returns the generated AST, or an
// error
func (t *Tokenizer) Finish() (AST, error) {
	if t._state != stateLiteral {
		return nil, UnexpectedEOFErr{
			idx:   t.idx,
			line:  t.line,
			state: t._state,
		}
	}

	t.commitLiteral()

	return cleanAST(t.ast), nil
}

// Tokenize takes all runes from the provided string, feeds an internal
// Tokenizer instance, and returns the result by calling Finish
func Tokenize(data string) (ast AST, err error) {
	t := NewTokenizer()
	for _, d := range []rune(data) {
		if err = t.Feed(d); err != nil {
			return
		}
	}
	return t.Finish()
}

func cleanAST(ast AST) AST {
	if ast == nil {
		return nil
	}
	var newAST AST
	for i, node := range ast {
		if i > 0 && i+1 < len(ast) && node.Kind() == KindLiteral {
			if node.(*Literal).String == "\n" &&
				ast[i-1].Kind() == KindConditional &&
				ast[i+1].Kind() == KindConditional {
				continue
			}
		}

		if node.Kind() == KindConditional {
			cond := node.(*Conditional)
			cond.Then = cleanAST(cond.Then)
			cond.Else = cleanAST(cond.Else)
			for i, elseIf := range cond.ElseIf {
				cond.ElseIf[i] = cleanAST(AST{elseIf})[0].(*Conditional)
			}
		}
		newAST = append(newAST, node)
	}
	return newAST
}
