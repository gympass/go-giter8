package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gympass/go-giter8/lexer"
	"github.com/gympass/go-giter8/props"
	"github.com/gympass/go-giter8/render"
)

func TestRendering(t *testing.T) {
	template := `$if(ok.truthy)$
OK!
$endif$
$if(notok.truthy)$
NOTOK
$endif$`

	p := props.FromMap(map[string]string{
		"ok":    "yes",
		"notok": "tchubaruba",
	})

	ast, err := lexer.Tokenize(template)
	require.NoError(t, err)

	exec := render.NewExecutor(p)
	r, err := exec.Exec(ast)
	require.NoError(t, err)
	assert.Equal(t, "\nOK!\n", r)
}

func TestNestedConditionals(t *testing.T) {
	template := `$if(parent.truthy)$
Parent OK
$if(child.truthy)$
Child OK
$endif$
$endif$`

	p := props.FromMap(map[string]string{
		"parent": "false",
		"child":  "true",
	})

	ast, err := lexer.Tokenize(template)
	require.NoError(t, err)

	exec := render.NewExecutor(p)
	r, err := exec.Exec(ast)
	require.NoError(t, err)
	assert.Equal(t, "", r)
}

func TestAbsentConditionalProperties(t *testing.T) {
	template := `$if(non-existing.truthy)$
$non-existing$
$endif$
$if(existing.truthy)$
Yay!
$endif$`

	p := props.FromMap(map[string]string{
		"existing": "true",
	})

	ast, err := lexer.Tokenize(template)
	require.NoError(t, err)

	exec := render.NewExecutor(p)
	r, err := exec.Exec(ast)
	require.NoError(t, err)
	assert.Equal(t, "\nYay!\n", r)

}

func TestPresentConditional(t *testing.T) {
	template := `$if(non-existing.present)$
$non-existing$
$endif$
$if(existing.present)$
foobar
$endif$`

	p := props.FromMap(map[string]string{
		"existing": "foobar",
	})

	ast, err := lexer.Tokenize(template)
	require.NoError(t, err)

	exec := render.NewExecutor(p)
	r, err := exec.Exec(ast)
	require.NoError(t, err)
	assert.Equal(t, "\nfoobar\n", r)

}
