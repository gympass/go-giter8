package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Gympass/go-giter8/lexer"
	"github.com/Gympass/go-giter8/props"
)

func TestRenderer(t *testing.T) {
	template := `$dashed-variable$`
	ast, err := lexer.Tokenize(template)
	require.NoError(t, err)
	prps := props.FromMap(map[string]string{
		"dashed-variable": "foo",
	})
	exec := NewExecutor(prps)
	res, err := exec.Exec(ast)
	require.NoError(t, err)
	assert.Equal(t, "foo", res)
}
