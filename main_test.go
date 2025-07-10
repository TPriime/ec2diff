package main

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tpriime/ec2diff/pkg"
	"github.com/tpriime/ec2diff/pkg/mocks"
	"github.com/tpriime/ec2diff/registry"
)

func TestRun_Args(t *testing.T) {
	t.Run("should show usage", func(t *testing.T) {
		var out bytes.Buffer
		err := run(t.Context(), []string{"-h"}, &out)

		assert.NoError(t, err)
		assert.Contains(t, out.String(), "Usage")
	})

	t.Run("should list attributes", func(t *testing.T) {
		var out bytes.Buffer
		err := run(t.Context(), []string{"-list-attributes"}, &out)

		assert.NoError(t, err)
		assert.Contains(t, out.String(), "Supported attributes")
		assert.Contains(t, out.String(), "instance_type")
	})

	t.Run("should invalidate attributes", func(t *testing.T) {
		var out bytes.Buffer
		err := run(t.Context(), []string{
			"-file", "test.tfstate",
			"-attrs", "unsupported_attr",
		}, &out)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not supported")
	})

	t.Run("should reject missing file and print usage", func(t *testing.T) {
		var out bytes.Buffer
		err := run(t.Context(), []string{"-instances", "i-123"}, &out)

		assert.Error(t, err)
		assert.Contains(t, out.String(), "Usage")
	})
}

func TestExecute_SuccessfulWithNoDrifts(t *testing.T) {
	state := pkg.InstanceMap{"i-abc": pkg.Instance{ID: "i-abc", State: "running"}}
	live := pkg.InstanceMap{"i-abc": pkg.Instance{ID: "i-abc", State: "running"}}

	parser := &mocks.MockParser{Parsed: state, Extensions: []string{".tfstate"}}
	fetcher := &mocks.MockLiveFetcher{Instances: live}
	printer := &mocks.MockReportPrinter{}

	cfg := &Config{
		FilePath:       "data.tfstate",
		HCLInstanceIDs: []string{"i-abc"},
		Attributes:     []string{pkg.AttrInstanceState},
		Registry:       registry.NewParserRegistry([]pkg.Parser{parser}),
		Fetcher:        fetcher,
		Checker:        &mocks.MockDriftChecker{},
		ReportPrinter:  printer,
		HelpFn:         func() {},
	}

	err := execute(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Len(t, printer.Output, 1)
	assert.Equal(t, "i-abc", printer.Output[0].InstanceID)
	assert.Empty(t, printer.Output[0].Drifts)
}

func TestExecute_MissingFile(t *testing.T) {
	called := false
	cfg := &Config{
		FilePath: "",
		HelpFn:   func() { called = true },
	}
	err := execute(context.Background(), cfg)
	assert.Error(t, err)
	assert.True(t, called)
}

func TestExecute_UnsupportedExtension(t *testing.T) {
	reg := registry.NewParserRegistry([]pkg.Parser{})
	cfg := &Config{
		FilePath: "unsupported.txt",
		Registry: reg,
	}
	err := execute(context.Background(), cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no parser found")
}

func TestExecute_ParseFails(t *testing.T) {
	parser := &mocks.MockParser{Err: errors.New("broken"), Extensions: []string{".tfstate"}}
	reg := registry.NewParserRegistry([]pkg.Parser{parser})
	cfg := &Config{
		FilePath: "file.tfstate",
		Registry: reg,
	}
	err := execute(context.Background(), cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestExecute_FetchFails(t *testing.T) {
	parser := &mocks.MockParser{Parsed: pkg.InstanceMap{}, Extensions: []string{".tfstate"}}
	fetcher := &mocks.MockLiveFetcher{Err: errors.New("timeout")}
	reg := registry.NewParserRegistry([]pkg.Parser{parser})

	cfg := &Config{
		FilePath: "file.tfstate",
		Registry: reg,
		Fetcher:  fetcher,
		HelpFn:   func() {},
	}

	err := execute(context.Background(), cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check drifts")
}

func TestParseCSV(t *testing.T) {
	input := "id1,id2 , id3"
	expected := []string{"id1", "id2", "id3"}
	actual := parseCommaSep(input)
	assert.Equal(t, expected, actual)
}

func TestValidateAttributes(t *testing.T) {
	err := validateAttributes([]string{"instance_type", "instance_state", "tags", "security_groups"})
	assert.NoError(t, err)

	err = validateAttributes([]string{"FakeAttr"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestSupportedAttributes(t *testing.T) {
	attrs := supportedAttributes()

	assert.Contains(t, attrs, pkg.AttrInstanceType, "supportedAttributes should contain 'instance_type'")
}
