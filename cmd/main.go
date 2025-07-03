package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/tpriime/ec2diff/pkg/aws"
	"github.com/tpriime/ec2diff/pkg/drift"
	"github.com/tpriime/ec2diff/pkg/terraform"
)

func main() {
	stateFile := flag.String("state", "", "Path to Terraform state JSON (`terraform show -json`)")
	region := flag.String("region", "us-east-1", "AWS region")
	instances := flag.String("instances", "", "Comma-separated instance IDs (defaults to all in state)")
	attrs := flag.String("attrs", "instance_type", "Comma-separated attributes to check")
	flag.Parse()

	if *stateFile == "" {
		fmt.Println("error: -state is required")
		os.Exit(1)
	}

	// parse desired attrs and state file
	wantedAttrs := splitTrim(*attrs)
	stateMap, err := terraform.ParseState(*stateFile)
	if err != nil {
		log.Fatalf("failed to parse state: %v", err)
	}

	// decide which IDs to process
	var ids []string
	if *instances == "" {
		for id := range stateMap {
			ids = append(ids, id)
		}
	} else {
		ids = splitTrim(*instances)
	}

	// AWS client
	awsClient, err := aws.NewClient(*region)
	if err != nil {
		log.Fatalf("failed to init AWS client: %v", err)
	}

	// drift detection concurrently
	var wg sync.WaitGroup
	results := make(chan drift.Report, len(ids))

	ctx := context.Background()
	for _, id := range ids {
		wg.Add(1)
		go func(instanceID string) {
			defer wg.Done()
			report := drift.CheckDrift(ctx, awsClient, stateMap, instanceID, wantedAttrs)
			results <- report
		}(id)
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

func splitTrim(in string) []string {
	fields := strings.Split(in, ",")
	var out []string
	for _, v := range fields {
		if t := strings.TrimSpace(v); t != "" {
			out = append(out, t)
		}
	}
	return out
}
