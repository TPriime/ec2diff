package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"
	"sync"

	"github.com/spf13/cobra"
	"github.com/tpriime/ec2diff/pkg"
	"github.com/tpriime/ec2diff/pkg/aws"
	"github.com/tpriime/ec2diff/pkg/drift"
	"github.com/tpriime/ec2diff/pkg/terraform"
)

var (
	stateFile   string
	region      string
	instanceIDs []string
	attrs       []string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ec2diff",
		Short: "Detect drift between Terraform state and AWS EC2 instances",
		Run:   run,
	}

	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		supportedAttr := pkg.SupportedAttributes()
		for _, attr := range attrs {
			if !slices.Contains(supportedAttr, attr) {
				log.Fatalf("attribute '%s' not supported. Supported attributes are %v", attr, supportedAttr)
				os.Exit(1)
			}
		}
	}

	rootCmd.Flags().StringVar(&stateFile, "state", "", "Path to Terraform state JSON (`terraform show -json`)")
	rootCmd.Flags().StringVar(&region, "region", "us-east-1", "AWS region")
	rootCmd.Flags().StringSliceVar(&instanceIDs, "instances", nil, "Comma-separated instance IDs (defaults to all in state)")
	rootCmd.Flags().StringSliceVar(&attrs, "attrs", nil, "Comma-separated attributes to check")
	rootCmd.MarkFlagRequired("state")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	// parse desired attrs and state file
	stateInstances, err := terraform.ParseState(stateFile)
	if err != nil {
		log.Fatalf("failed to parse state: %v", err)
	}

	// decide which IDs to process
	selectedInstances := []pkg.Instance{}
	if len(instanceIDs) != 0 {
		for _, i := range instanceIDs {
			if inst, ok := stateInstances[i]; ok {
				selectedInstances = append(selectedInstances, inst)
			} else {
				log.Fatalf("instance not found in terraform state, instaneID: %s", i)
			}
		}
	} else {
		for _, inst := range stateInstances {
			selectedInstances = append(selectedInstances, inst)
		}
	}

	// AWS client
	awsClient, err := aws.NewClient(region)
	if err != nil {
		log.Fatalf("failed to init AWS client: %v", err)
	}

	// drift detection concurrently
	var wg sync.WaitGroup
	results := make(chan drift.Report, len(selectedInstances))

	ctx := context.Background()
	for _, instance := range selectedInstances {
		wg.Add(1)
		go func(stateInstance pkg.Instance) {
			defer wg.Done()
			awsIntance, err := awsClient.GetInstance(ctx, stateInstance.ID)
			if err != nil {
				if errors.Is(err, aws.ErrNotFound) {
					awsIntance = &pkg.Instance{} // instance is deleted, use empty object
				} else {
					log.Fatalf("failed to fetch instance from AWS: %v", err)
				}
			}

			report := drift.CheckDrift(ctx, stateInstance, *awsIntance, attrs)
			results <- report
		}(instance)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// collect and output JSON
	all := []drift.Report{}
	for r := range results {
		all = append(all, r)
	}
	out, _ := json.MarshalIndent(all, "", "  ")
	fmt.Println(string(out))
}
