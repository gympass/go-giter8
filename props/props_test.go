package props

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPropsParsing(t *testing.T) {
	props := `name=Project Name
nameUpperSnake=$name;format="upper,snake"$
normalized=$name;format="normalize"$
organization=com.foo
package=$normalized;format="word"$
minimumCoverage=10
descriptions=Project descriptions
namespace=foo
productionCluster=production
dashed-variable=value
`
	allProps, err := ParseProperties(props)
	require.NoError(t, err)
	assert.Equal(t, "Project Name", allProps.MustGet("name"))
	assert.Equal(t, "$name;format=\"upper,snake\"$", allProps.MustGet("nameUpperSnake"))
	assert.Equal(t, "$name;format=\"normalize\"$", allProps.MustGet("normalized"))
	assert.Equal(t, "com.foo", allProps.MustGet("organization"))
	assert.Equal(t, "$normalized;format=\"word\"$", allProps.MustGet("package"))
	assert.Equal(t, "10", allProps.MustGet("minimumCoverage"))
	assert.Equal(t, "foo", allProps.MustGet("namespace"))
	assert.Equal(t, "production", allProps.MustGet("productionCluster"))
	assert.Equal(t, "value", allProps.MustGet("dashed-variable"))
}
