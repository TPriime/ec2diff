package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/tpriime/ec2diff/pkg"
	"github.com/tpriime/ec2diff/pkg/logger"
)

// awsFetcher fetches EC2 instances from AWS.
type awsFetcher struct {
	client    ec2API
	pageLimit int32
}

// ec2API defines the subset of EC2 client methods used.
// This allows mocking the EC2 API for testing.
type ec2API interface {
	DescribeInstances(
		ctx context.Context,
		params *ec2.DescribeInstancesInput,
		optFns ...func(*ec2.Options),
	) (*ec2.DescribeInstancesOutput, error)
}

// NewAwsFetcher initializes an AWS EC2 client and returns a LiveFetcher.
func NewAwsFetcher(ctx context.Context, pageLimit int32) (pkg.PaginatedLiveFetcher, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	return &awsFetcher{client: ec2.NewFromConfig(cfg), pageLimit: pageLimit}, nil
}

// Fetch retrieves all EC2 instances from AWS in a paginated manner and maps them by instance ID.
// - onPageFn declares function to run per pagination
func (f *awsFetcher) Fetch(ctx context.Context, onPageFn func(page int, instances pkg.InstanceMap) bool) error {
	paginator := ec2.NewDescribeInstancesPaginator(f.client, &ec2.DescribeInstancesInput{MaxResults: &f.pageLimit})
	pageCount := 1

	for paginator.HasMorePages() {
		logger.Info(ctx, "Fetching next batch of aws instances...", "op", "awsFetcher.Fetch")
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch page %d: %w", pageCount, err)
		}

		instances := make(pkg.InstanceMap)
		for _, reservation := range page.Reservations {
			for _, instance := range reservation.Instances {
				// Convert AWS instance to local model and store
				instances[*instance.InstanceId] = toModel(instance)
			}
		}

		onPageFn(pageCount, instances)
		pageCount++
	}

	return nil
}

// toModel maps an AWS EC2 instance to the local pkg.Instance type.
func toModel(inst types.Instance) pkg.Instance {
	tags := map[string]string{}
	for _, t := range inst.Tags {
		tags[valstr(t.Key)] = valstr(t.Value)
	}

	var state string
	if inst.State != nil {
		state = string(inst.State.Name)
	}

	sgs := []string{}
	for _, sg := range inst.SecurityGroups {
		sgs = append(sgs, valstr(sg.GroupName))
	}

	return pkg.Instance{
		ID:             valstr(inst.InstanceId),
		Type:           string(inst.InstanceType),
		State:          state,
		KeyName:        valstr(inst.KeyName),
		Tags:           tags,
		SecurityGroups: sgs,
		PublicIP:       valstr(inst.PublicIpAddress),
	}
}

// valstr safely dereferences a *string to a string.
func valstr(ptr *string) (s string) {
	if ptr != nil {
		s = *ptr
	}
	return
}
