package render

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Ref: github.com/Gympass/josie/issues/57
func TestIssueJosie57(t *testing.T) {
	name := "Project Name"
	nameWithSpace := wordSpace(name)
	fullPackage := upperCamel(nameWithSpace)
	packaged := packageNaming(fullPackage)
	assert.Equal(t, "ProjectName", packaged)
}
