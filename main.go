package main

import (
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/spf13/cobra"
	"github.com/tpriime/ec2diff/pkg"
	"github.com/tpriime/ec2diff/pkg/aws"
	"github.com/tpriime/ec2diff/pkg/drift"
	"github.com/tpriime/ec2diff/pkg/terraform"
)

type input struct {
	stateFile   string
	hclFile     string
	region      string
	instanceIDs []string
	attrs       []string
}

type options struct {
	client pkg.Client
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

			// parse state or hcl file
			fileType := "state"
			if in.stateFile != "" {
				stateInstances, err = terraform.ParseState(in.stateFile)
			} else {
				stateInstances, err = terraform.ParseHCL(in.hclFile, in.instanceIDs)
				fileType = "hcl"
			}
			if err != nil {
				log.Fatalf("failed to parse %s file: %v", fileType, err)
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

	cmd.Flags().StringVar(&in.stateFile, "state", "", "Path to Terraform state JSON (`terraform show -json`)")
	cmd.Flags().StringVar(&in.hclFile, "hcl", "", "Path to HCL definition file (.hcl)")
	cmd.Flags().StringVar(&in.region, "region", "us-east-1", "AWS region")
	cmd.Flags().StringSliceVar(&in.instanceIDs, "instances", nil, "Comma-separated instance IDs (defaults to all in state)")
	cmd.Flags().StringSliceVar(&in.attrs, "attrs", nil, "Comma-separated attributes to check")

	cmd.MarkFlagsOneRequired("state", "hcl")
	cmd.MarkFlagsMutuallyExclusive("state", "hcl")

	return cmd
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
