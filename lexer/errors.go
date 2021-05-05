package lexer

import "fmt"

type UnexpectedEOFErr struct {
	idx   int
	state state
	line  int
}

func (u UnexpectedEOFErr) Error() string {
	return fmt.Sprintf("Unexpected end of input at line %d (index %d). Tokenizer state was %s", u.line+1, u.idx, u.state)
}

type UnexpectedLinebreakErr struct {
	idx  int
	line int
}

func (u UnexpectedLinebreakErr) Error() string {
	return fmt.Sprintf("Unexpected linebreak at line %d (index %d)", u.line+1, u.idx)
}

type UnexpectedTokenErr struct {
	idx   int
	token string
	line  int
}

func (u UnexpectedTokenErr) Error() string {
	return fmt.Sprintf("Unexpected token `%s' at line %d (index %d)", u.token, u.line+1, u.idx)
}

type UnsupportedConditionalHelperErr struct {
	idx    int
	helper string
	line   int
}

func (u UnsupportedConditionalHelperErr) Error() string {
	return fmt.Sprintf("Unsupported conditional helper `%s' at line %d (index %d)", u.helper, u.line+1, u.idx)
}

type InvalidConditionalExpressionErr struct {
	idx  int
	line int
	expr string
}

func (u InvalidConditionalExpressionErr) Error() string {
	return fmt.Sprintf("Invalid conditional expression `%s' at line %d (index %d)", u.expr, u.line+1, u.idx)
}
