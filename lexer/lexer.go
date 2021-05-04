package lexer

import (
	"strings"
	"unicode"

	"github.com/Gympass/go-giter8/sb"
)

type Literal string
type Template struct {
	Name    string
	Options map[string]string
}

type AST []interface{}

type state int

const (
	stateLiteral state = iota
	stateTemplateName
	stateTemplateOptionName
	stateTemplateOptionValueBegin
	stateTemplateOptionValue
	stateTemplateOptionOrEnd
)

const ESCAPE = rune('\\')
const DELIM = rune('$')
const NEWLINE = rune('\n')
const SEMICOLON = rune(';')
const EQUALS = rune('=')
const QUOT = rune('"')
const COMMA = rune(',')
const SPACE = rune(' ')
const HTAB = rune('\t')

func isSpace(r rune) bool {
	return r == SPACE || r == HTAB
}

type Tokenizer struct {
	ast AST
	tmp *sb.StringBuilder

	templateName    *sb.StringBuilder
	optionName      *sb.StringBuilder
	optionValue     *sb.StringBuilder
	templateOptions map[string]string
	state           state

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
		state:           0,
		lastFedRune:     0,
		idx:             0,
		line:            0,
	}
}

func (t *Tokenizer) commitLiteral() {
	if t.tmp.Len() == 0 {
		return
	}
	t.ast = append(t.ast, Literal(t.tmp.String()))
	t.tmp.Reset()
}

func (t *Tokenizer) commitTemplate() {
	if t.templateName.Len() == 0 {
		return
	}
	t.ast = append(t.ast, Template{
		Name:    strings.TrimSpace(t.templateName.String()),
		Options: t.templateOptions,
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

func (t *Tokenizer) lastRune() rune {
	return t.lastFedRune
}

func (t Tokenizer) unexpectedToken(token rune) error {
	return UnexpectedTokenErr{idx: t.idx, line: t.line, token: string(token)}
}

func (t Tokenizer) unexpectedLineBreak() error {
	return UnexpectedLinebreakErr{line: t.line, idx: t.idx}
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
	switch t.state {
	case stateLiteral:
		if chr == DELIM && t.lastRune() != ESCAPE {
			t.commitLiteral()
			t.state = stateTemplateName
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
			t.commitTemplate()
			t.state = stateLiteral
			return nil
		} else if chr == SPACE {
			return t.unexpectedToken(SPACE)
		} else if chr == SEMICOLON {
			t.state = stateTemplateOptionName
			return nil
		} else if chr == NEWLINE {
			return t.unexpectedLineBreak()
		}
		if t.templateName.Len() == 0 && !unicode.IsLetter(chr) {
			return t.unexpectedToken(chr)
		}
		if !unicode.IsLetter(chr) && !unicode.IsDigit(chr) {
			return t.unexpectedToken(chr)
		}
		t.templateName.WriteRune(chr)

	case stateTemplateOptionName:
		if chr == DELIM {
			if t.templateName.Len() == 0 {
				return t.unexpectedToken(DELIM)
			}
			t.commitTemplate()
			t.state = stateLiteral
			return nil
		} else if chr == EQUALS {
			t.state = stateTemplateOptionValueBegin
			return nil
		}
		t.optionName.WriteRune(chr)

	case stateTemplateOptionValueBegin:
		if isSpace(chr) {
			return nil
		} else if chr == QUOT {
			t.state = stateTemplateOptionValue
			return nil
		}
		return t.unexpectedToken(chr)

	case stateTemplateOptionValue:
		if chr == QUOT && t.lastRune() != ESCAPE {
			t.state = stateTemplateOptionOrEnd
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
			t.state = stateTemplateOptionName
			return nil
		} else if chr == DELIM {
			t.state = stateLiteral
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
	if t.state != stateLiteral {
		return nil, UnexpectedEOFErr{
			idx:   t.idx,
			line:  t.line,
			state: t.state,
		}
	}

	t.commitLiteral()

	return t.ast, nil
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
