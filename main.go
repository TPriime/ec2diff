package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"
	"github.com/tpriime/ec2diff/pkg"
	"github.com/tpriime/ec2diff/pkg/aws"
	"github.com/tpriime/ec2diff/pkg/drift"
	"github.com/tpriime/ec2diff/pkg/hcl"
	"github.com/tpriime/ec2diff/pkg/tfstate"
)

type input struct {
	file        string
	region      string
	instanceIDs []string
	attrs       []string
}

type options struct {
	client pkg.LiveFetcher
}

func main() {
	rootCmd := setupCommand(options{})
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func setupCommand(opts options) *cobra.Command {
	in := input{}
	cmd := &cobra.Command{
		Use:   "ec2diff",
		Short: "Detect drift between Terraform state and AWS EC2 instances",
		Run: func(cmd *cobra.Command, _ []string) {

			// setup client, allows mocking
			client := opts.client
			if client == nil {
				var err error
				client, err = aws.NewClient(in.region)
				if err != nil {
					log.Fatalf("failed to init AWS client: %v", err)
				}
			}

			var stateInstances map[string]pkg.Instance
			var err error

			// dynamically parse file based on file extention
			ext := filepath.Ext(in.file)
			for _, p := range parsers() {
				if slices.Contains(p.SupportedTypes(), ext) {
					stateInstances, err = p.Parse(in.file, in.instanceIDs)
					if err != nil {
						log.Fatalf("failed to parse file: %v", err)
					}
					break
				}
			}

			// filter instances to only supplied ids
			targets, err := getTargetInstances(stateInstances, in.instanceIDs)
			if err != nil {
				log.Fatalf("failed to filter instances: %v", err)
			}

			// main checker
			reports := drift.CheckDrift(client, targets, in.attrs)

			// print report
			for _, r := range reports {
				r.Print(cmd.OutOrStdout())
			}
		},
	}

	cmd.PreRun = func(_ *cobra.Command, _ []string) {
		supportedAttr := pkg.SupportedAttributes()
		for _, attr := range in.attrs {
			if !slices.Contains(supportedAttr, attr) {
				log.Fatalf("attribute '%s' not supported. Supported attributes are %v", attr, supportedAttr)
				os.Exit(1)
			}
		}
	}

	// Add list-attributes subcommand
	cmd.AddCommand(&cobra.Command{
		Use:   "list-attributes",
		Short: "List supported attributes for drift detection",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Supported attributes:")
			for _, attr := range pkg.SupportedAttributes() {
				fmt.Println(" -", attr)
			}
		},
	})

	cmd.Flags().StringVar(&in.file, "file", "", "Path to file (.hcl, .tfstate)")
	cmd.Flags().StringVar(&in.region, "region", "us-east-1", "AWS region")
	cmd.Flags().StringSliceVar(&in.instanceIDs, "instances", nil, "Comma-separated instance IDs (defaults to all in state)")
	cmd.Flags().StringSliceVar(&in.attrs, "attrs", nil, "Comma-separated attributes to check")

	if err := cmd.MarkFlagRequired("file"); err != nil {
		log.Fatal("Must provide file path")
	}

	return cmd
}

func parsers() []pkg.Parser {
	return []pkg.Parser{
		tfstate.NewTfStateParser(),
		hcl.NewHclParser(),
	}
}

// getTargetInstances returns a slice of pkg.Instance corresponding to the provided instance IDs.
// If no IDs are specified, it returns all instances from the given stateInstances map.
// Returns an error if any of the specified IDs are not found in the stateInstances map.
func getTargetInstances(stateInstances map[string]pkg.Instance, ids []string) ([]pkg.Instance, error) {
	selectedInstances := []pkg.Instance{}

	// select all instances if none is specified
	if len(ids) == 0 {
		for _, inst := range stateInstances {
			selectedInstances = append(selectedInstances, inst)
		}
		return selectedInstances, nil
	}

	for _, id := range ids {
		if inst, ok := stateInstances[id]; ok {
			selectedInstances = append(selectedInstances, inst)
		} else {
			return nil, fmt.Errorf("instance not found in terraform state, instaneID: %s", id)
		}
	}

	return selectedInstances, nil
}
