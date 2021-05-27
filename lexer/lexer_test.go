package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLexerLiteral(t *testing.T) {
	_, err := Tokenize("This is basically a 'big' literal.")
	require.NoError(t, err)
}

func TestLexerSimpleTemplate(t *testing.T) {
	_, err := Tokenize("$simpleTemplate$")
	require.NoError(t, err)
}

func TestLexerTemplateWithOption(t *testing.T) {
	_, err := Tokenize("$simpleTemplate; format=\"test, foo, bar\"$")
	require.NoError(t, err)
}

func TestLexerTemplateWithOptions(t *testing.T) {
	_, err := Tokenize("$simpleTemplate; format=\"test, foo, bar\", foo = \"bar\"$")
	require.NoError(t, err)
}

func TestBrokenTemplate(t *testing.T) {
	template := "hello, $world;foo=\"\n$\""
	_, err := Tokenize(template)
	require.Error(t, err)
}

func TestEscapes(t *testing.T) {
	template := "hello, \\$world $foo;foo=\"\\\"<-quote\"$\""
	_, err := Tokenize(template)
	require.NoError(t, err)
}

func TestEscapes01(t *testing.T) {
	template := "RUN echo \"\\${SSH_PRIVATE_KEY}\" > /root/.ssh/id_rsa"
	_, err := Tokenize(template)
	require.NoError(t, err)
}

func TestConditional(t *testing.T) {
	template := "$if(foobar.truthy)$foo$endif$"
	_, err := Tokenize(template)
	require.NoError(t, err)
}

func TestConditionalElse(t *testing.T) {
	template := "$if(foobar.truthy)$foo$else$bar$endif$"
	_, err := Tokenize(template)
	require.NoError(t, err)
}

func TestConditionalElseIf(t *testing.T) {
	template := `$if(foobar.truthy)$
foo
$elseif(foobar.truthy)$
bar
$else$
baz
$endif$`
	_, err := Tokenize(template)
	require.NoError(t, err)
}

func TestConditionalUnorderedStructure(t *testing.T) {
	template := `$if(foobar.truthy)$
foo
$else$
bar
$elseif(foobar.truthy)$
baz
$endif$`
	_, err := Tokenize(template)
	require.Error(t, err)
}

func TestCompositeFormatting(t *testing.T) {
	template := `$name__decap$`
	ast, err := Tokenize(template)
	require.NoError(t, err)
	require.Equal(t, 1, len(ast))
	node := ast[0]
	assert.Equal(t, KindTemplate, node.Kind())
	tmp := node.(*Template)
	assert.Equal(t, "name", tmp.Name)
	assert.Equal(t, 1, len(tmp.Options))
	fmt, ok := tmp.Options["format"]
	assert.True(t, ok)
	assert.Equal(t, "decap", fmt)
}

func TestDashedVariables(t *testing.T) {
	template := "$some-variable$"
	ast, err := Tokenize(template)
	require.NoError(t, err)
	require.Equal(t, 1, len(ast))
	node := ast[0]
	assert.Equal(t, KindTemplate, node.Kind())
	tmp := node.(*Template)
	assert.Equal(t, "some-variable", tmp.Name)
}

func TestAST_IsPureLiteral(t *testing.T) {
	template := "A pure-literal \\$value"
	ast, err := Tokenize(template)
	require.NoError(t, err)
	assert.True(t, ast.IsPureLiteral())
}
