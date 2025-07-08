package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tpriime/ec2diff/pkg"
	"github.com/tpriime/ec2diff/pkg/mocks"
)

func TestNewParserRegistry_SingleParser(t *testing.T) {
	parser := &mocks.MockParser{Extensions: []string{".tfstate", ".json"}}
	reg := NewParserRegistry([]pkg.Parser{parser})

	for _, ext := range parser.Extensions {
		p, ok := reg.Get("file" + ext)
		assert.True(t, ok, "expected parser to be found for extension: %s", ext)
		assert.Equal(t, parser, p)
	}
}

func TestNewParserRegistry_UnknownExtension(t *testing.T) {
	parser := &mocks.MockParser{Extensions: []string{".tfstate"}}
	reg := NewParserRegistry([]pkg.Parser{parser})

	p, ok := reg.Get("main.yaml")
	assert.False(t, ok)
	assert.Nil(t, p)
}

func TestNewParserRegistry_MultipleParsers(t *testing.T) {
	tf := &mocks.MockParser{Extensions: []string{".tfstate"}}
	hcl := &mocks.MockParser{Extensions: []string{".hcl"}}

	reg := NewParserRegistry([]pkg.Parser{tf, hcl})

	p1, ok1 := reg.Get("a.tfstate")
	p2, ok2 := reg.Get("b.hcl")

	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.Equal(t, tf, p1)
	assert.Equal(t, hcl, p2)
}
