package operations

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeGoType(t *testing.T) {
	assert.Equal(t, "map[]User", normalizeGoType("map[string]User"))
	assert.Equal(t, "map[]pkg.User", normalizeGoType("map[int]pkg.User"))
	assert.Equal(t, "[]map[]pkg.User", normalizeGoType("[]map[string]pkg.User"))
	assert.Equal(t, "[]pkg.User", normalizeGoType(" []*pkg.User "))
	assert.Equal(t, "pkg.User", normalizeGoType("*pkg.User"))
}

func TestParseResponseFields_WithGenericMappings(t *testing.T) {
	parsed, err := parseResponseFields(`200 {object} yhttp.DataRes{data= []*feedback.Item, total=int32, extra=map[string]feedback.Meta} "ok"`)

	require.NoError(t, err)
	assert.Equal(t, "200", parsed.status)
	assert.Equal(t, "{object}", parsed.jsonType)
	assert.Equal(t, `yhttp.DataRes{data=[]feedback.Item,total=int32,extra=map[]feedback.Meta}`, parsed.goType)
	assert.Equal(t, "ok", parsed.description)
}

func TestParseResponseFields_DescriptionOnly(t *testing.T) {
	parsed, err := parseResponseFields(`204 "No content"`)

	require.NoError(t, err)
	assert.Equal(t, "204", parsed.status)
	assert.Empty(t, parsed.jsonType)
	assert.Empty(t, parsed.goType)
	assert.Equal(t, "No content", parsed.description)
}
