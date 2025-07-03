# ec2diff

`ec2diff` is a command-line tool written in Go for comparing the state of AWS EC2 instances between two points in time or between two AWS accounts/regions. It helps DevOps engineers and cloud administrators quickly identify changes in EC2 infrastructure, such as instance launches, terminations, or modifications.

## Features

- List EC2 instances in a given AWS account/region.
- Compare EC2 instance states between two snapshots or environments.
- Output differences in a human-readable format.
- Supports filtering by tags, instance states, or other attributes.

## Installation

1. **Clone the repository:**
   ```sh
   git clone https://github.com/yourusername/ec2diff.git
   cd ec2diff
   ```

2. **Build the binary:**
   ```sh
   go build -o ec2diff
   ```

3. **(Optional) Install globally:**
   ```sh
   sudo mv ec2diff /usr/local/bin/
   ```

## Usage

### Basic Usage

```sh
./ec2diff --region us-east-1 snapshot1.json snapshot2.json
```

- `--region`: AWS region to query (e.g., `us-east-1`).
- `snapshot1.json` and `snapshot2.json`: JSON files containing EC2 instance snapshots.

### Creating Snapshots

To create a snapshot of your current EC2 instances:

```sh
./ec2diff snapshot --region us-east-1 --output snapshot1.json
```

### Comparing Snapshots

To compare two snapshots and see the differences:

```sh
./ec2diff diff snapshot1.json snapshot2.json
```

### Filtering by Tag

```sh
./ec2diff diff --tag "Environment=prod" snapshot1.json snapshot2.json
```

## Example Output

```
+ i-0123456789abcdef0   (new instance)
- i-0fedcba9876543210   (terminated)
~ i-0a1b2c3d4e5f6g7h8   (modified: instance type t2.micro -> t3.micro)
```

## Limitations

- Only supports EC2 instances (no support for other AWS resources).
- Requires AWS credentials with sufficient permissions.
- Snapshot files must be generated using the tool itself.
- Does not support live, real-time diffing between accounts/regions (must use snapshots).
- Limited filtering options (currently by tag and state).

## Extensions & Future Work

- Support for other AWS resources (e.g., EBS volumes, security groups).
- Real-time diffing between two AWS accounts/regions without snapshots.
- Output in additional formats (YAML, CSV, HTML).
- Integration with CI/CD pipelines for automated drift detection.
- More advanced filtering and query capabilities.
- Web UI for visualizing differences.

## Contributing

Contributions are welcome! Please open issues or pull requests on GitHub.

## License

MIT License. See [LICENSE](LICENSE) for details.

## Authors

- [Your Name](https://github.com/yourusername)

