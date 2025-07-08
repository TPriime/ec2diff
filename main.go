package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/tpriime/ec2diff/pkg"
	"github.com/tpriime/ec2diff/pkg/aws"
	"github.com/tpriime/ec2diff/pkg/drift"
	"github.com/tpriime/ec2diff/pkg/hcl"
	"github.com/tpriime/ec2diff/pkg/tableprinter"
	"github.com/tpriime/ec2diff/pkg/tfstate"
	"github.com/tpriime/ec2diff/registry"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// Config holds parsed inputs and injected dependencies for drift checking.
type Config struct {
	// CLI args
	FilePath    string   // Path to HCL or tfstate file
	InstanceIDs []string // EC2 instance IDs to check
	Attributes  []string // EC2 attributes to compare
	ShowHelp    bool     // Whether to display CLI help
	ListAttrs   bool     // Whether to list supported attributes

	// Dependencies
	Writer        io.Writer
	Registry      *registry.ParserRegistry
	Fetcher       pkg.LiveFetcher
	Checker       pkg.DriftChecker
	ReportPrinter pkg.ReportPrinter
	HelpFn        func()
}

// run parses flags and injects default dependencies before executing logic.
func run(args []string, out io.Writer) error {
	cfg, err := parseFlags(args, out)
	if err != nil {
		return err
	}

	// Show help and exit
	if cfg.ShowHelp {
		cfg.HelpFn()
		return nil
	}

	// List attributes and exit
	if cfg.ListAttrs {
		fmt.Fprintln(cfg.Writer, "Supported attributes:")
		for _, attr := range supportedAttributes() {
			fmt.Fprintln(cfg.Writer, " -", attr)
		}
		return nil
	}

	// Initialize dependencies
	ctx := context.Background()
	cfg.Registry = registry.NewParserRegistry([]pkg.Parser{
		tfstate.NewTfStateParser(),
		hcl.NewHclParser(),
	})
	cfg.Fetcher, err = aws.NewAwsFetcher(ctx)
	if err != nil {
		return fmt.Errorf("failed to init AWS client: %w", err)
	}
	cfg.Checker = drift.NewDriftChecker()
	cfg.ReportPrinter = tableprinter.NewTablePrinter(out)

	return execute(ctx, cfg)
}

// parseFlags reads command-line arguments and returns a populated Config.
func parseFlags(args []string, out io.Writer) (*Config, error) {
	fs := flag.NewFlagSet("ec2diff", flag.ContinueOnError)
	fs.SetOutput(out)

	file := fs.String("file", "", "Path to file (.hcl or .tfstate)")
	instances := fs.String("instances", "", "Comma-separated instance IDs")
	attrs := fs.String("attrs", "", "Comma-separated attributes to check")
	listAttrs := fs.Bool("list-attributes", false, "List supported attributes")
	showHelp := fs.Bool("h", false, "Show help")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	cfg := &Config{
		FilePath:    *file,
		InstanceIDs: parseCommaSep(*instances),
		Attributes:  parseCommaSep(*attrs),
		ListAttrs:   *listAttrs,
		ShowHelp:    *showHelp,
		Writer:      out,
		HelpFn:      fs.Usage,
	}

	return cfg, nil
}

// execute performs the main comparison logic based on the provided Config.
func execute(ctx context.Context, cfg *Config) error {
	if cfg.FilePath == "" {
		cfg.HelpFn()
		return errors.New("missing required -file argument")
	}

	if err := validateAttributes(cfg.Attributes); err != nil {
		return err
	}

	parser, ok := cfg.Registry.Get(cfg.FilePath)
	if !ok {
		return fmt.Errorf("no parser found for file extension %s", filepath.Ext(cfg.FilePath))
	}

	// Parse local state
	state, err := parser.Parse(cfg.FilePath, cfg.InstanceIDs)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// Fetch live ec2 resources
	live, err := cfg.Fetcher.Fetch(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch live EC2 instances: %w", err)
	}

	// Use all supported attributes if none are specified.
	if len(cfg.Attributes) == 0 {
		cfg.Attributes = supportedAttributes()
	}

	// Check for drifts and report
	reports := cfg.Checker.CheckDrift(ctx, live, state, cfg.Attributes)
	cfg.ReportPrinter.Print(reports)

	return nil
}

// parseCommaSep splits a comma-separated string into a clean string slice.
func parseCommaSep(input string) []string {
	if input == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	var clean []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			clean = append(clean, p)
		}
	}
	return clean
}

// validateAttributes ensures each input attribute is supported.
func validateAttributes(attrs []string) error {
	if len(attrs) == 0 {
		return nil
	}
	supported := supportedAttributes()
	for _, attr := range attrs {
		if !slices.Contains(supported, attr) {
			return fmt.Errorf("attribute '%s' not supported. Supported attributes: %v", attr, supported)
		}
	}
	return nil
}

func supportedAttributes() []string {
	return []string{
		pkg.AttrInstanceType,
		pkg.AttrInstanceState,
		pkg.AttrKeyName,
		pkg.AttrTags,
		pkg.AttrSecurityGroups,
		pkg.AttrPublicIP,
	}
}
