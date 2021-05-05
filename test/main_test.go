package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Gympass/go-giter8/lexer"
	"github.com/Gympass/go-giter8/props"
	"github.com/Gympass/go-giter8/render"
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
