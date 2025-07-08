# ec2diff

`ec2diff` is a command-line tool written in Go for comparing the state of AWS EC2 instances with terraform state or hcl file. It helps DevOps engineers and cloud administrators quickly identify changes in EC2 infrastructure, such as instance launches, terminations, or modifications.

## Features

- Compare EC2 instance states between aws and the give terraform state or hcl file.
- Output differences in a human-readable format.
- Supports filtering by tags, instance states, or other attributes.

## Installation

1. **Clone the repository:**
   ```sh
   git clone https://github.com/tpriime/ec2diff.git
   cd ec2diff
   ```

2. **Build the binary:**
   ```sh
   go build -o ec2diff
   ```

   Or, using Docker:
   ```sh
   docker build -t ec2diff .
   ```

## Usage


### 🔐 AWS Authentication Setup

This tool uses the **AWS SDK for Go v2**, which follows the standard AWS CLI authentication methods.

Authenticate by setting the following env variables in your terminal:

```bash
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_REGION=your-aws-region
```

or by setting the equivalent profile:
```bash
export AWS_PROFILE=your-profile
```

### Basic Usage
```sh
./ec2diff --file ./examples/resources/terraform.tfstate 
```

Or, using Docker:

```sh
docker run --rm -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY \
   -v $(pwd):/data ec2diff \
   --file /data/examples/resources/terraform.tfstate 
```

To get a list of supported attributes run:
```sh
./ec2diff --list-attributes
```

You can see a list of available arguments by running help:
```sh 
./ec2diff -h
```

### Advanced Usage Example

Compare multiple EC2 instances and additional attributes using a Terraform state file:

```sh
./ec2diff --file ./examples/resources/instance.hcl \
   --attrs="instance_type,instance_state,security_groups" \
   --instances="i-0eb39d79613c9e43a,i-0bb19d79513a9e490,i-0ab19d89513c8e3ac"
```

## Example Output

The following output indicate that 3 instances exist live, in this case, AWS: 1 instance
is missing in the state file, 1 has drifted, and 1 has no drifts:

```
==============================
            REPORT
==============================

Instance [1]      : i-09f95c75f6cea3357
Comment           : Missing state

Attribute         Live                                  State
-------------     ----------------------------------    ------------------------------
instance_type     t4g.micro                             -                             
instance_state    running                               -                             
key_name          test                                  -                             
tags              {"Name":"driftcheck-missing"}         -                             
security_groups   ["default","launch-wizard-1"]         -                             
public_ip         54.237.208.142                        -                             

—

Instance [2]      : i-0eb39d79613c9e43a
Comment           : Drifts detected

Attribute         Live                                   State
-------------     ----------------------------------     ------------------------------
instance_state    stopped                                running                       
tags              {"Env":"Dev","Name":"DriftEC2"}        {"Name":"DriftEC2"}           
public_ip                                                3.80.95.115                   

—

Instance [3]      : i-0022023
Comment           : No drifts detected
```

---

## Design Decisions & Tradeoffs

### Structure
```sh
├── examples
│   ├── resources/
│   └── terraform/
├── pkg
│   ├── aws/
│   ├── drift/
│   ├── hcl/
│   ├── mocks/
│   ├── tableprinter/
│   ├── tfstate/
│   ├── driftchecker.go
│   ├── instance.go
│   ├── livefetcher.go
│   ├── parser.go
│   └── reportprinter.go
├── registry
│   └── parser_registry.go
└── main.go
```

- [**examples**](./examples) contain sample [**resources**](./examples/resources) that could be used as input to the program. It also contains a sample [**terraform**](./examples/terraform) code that could be run to setup an EC2 instance on AWS.
- [**pkg**](./pkg) is where most interface definitions for the application are defined and implemented. Notably:
   - [`LiveFetcher`](./pkg/livefetcher.go) interface for fetching instances from a live source (e.g. AWS).
   - [`Parser`](./pkg/parser.go) interface for parsing state files passed to the program to extract instance definitions.
   - [`DriftChecker`](./pkg/driftchecker.go) interface abstracts logic for comparing instances to detect differences/drifts.
   - [`ReportPrinter`](./pkg/reportprinter.go) interface abstracts logic for presenting/printing reports of drifts.
- [**registry**](./registry) registers available parsers.
- [main.go](./main.go) the program's entry point.



### Language & Libraries
- **Go** was chosen for its strong concurrency support, static typing, and suitability for CLI tools.
- **Cobra** provides a robust CLI framework, making argument parsing and help generation straightforward.
- **AWS SDK for Go v2** is used for modern, efficient AWS API access.
- **HCL (HashiCorp Configuration Language)**: For parsing infrastructure definitions.
- **go-cmp** is used for deep comparison of attributes, offering more flexibility and safety than `reflect.DeepEqual`.

### State Comparison Approach
- The tool compares the desired state (from Terraform state or HCL) with the actual AWS state.
- Attribute comparison is attribute-driven, allowing users to specify which fields to check, improving flexibility and performance.
- The common state interface ([`pkg.Instance`](./pkg/instance.go)) also allows for easy extensibility of new supported attribute fields and file formats.

### Concurrency
- Drift detection is performed concurrently for each instance using goroutines and a WaitGroup, improving performance for large infrastructures.

### Error Handling
- Errors are handled with clear log messages and immediate exit to avoid partial or misleading results.
- If an instance is missing in AWS, it is treated as deleted and compared as an empty object.

### Tradeoffs & Limitations
- **Concurrency**: The tool currently uses a simple, unbounded concurrency group, which can consume significant resources for very large instance lists, but is convenient for simple workloads.
- **Nesting**: Supporting deeply nested attributes may require additional effort, as reflection-based comparison has been avoided for performance reasons.

### Extensibility
- The modular design (separate packages for AWS, Terraform, and drift detection) makes it easier to add support for new resource types or input formats in the future.
- The direct use of an attribute list makes it easier to support more future attributes

---

## Future Improvements
* Accept aws response json as input
* Export drift reports in formats such as JSON, HTML, or CSV