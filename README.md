# ec2diff

`ec2diff` is a command-line tool written in Go for comparing the state of AWS EC2 instances with a Terraform state file. It helps DevOps engineers and cloud administrators quickly identify changes in EC2 infrastructure, such as instance launches, terminations, or modifications. It is also useful during migrations.

## Features

- Compare EC2 instance states between aws and the given terraform state file.
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

---

#### Using Localstack

To use localstack, set the following env variables:
```bash
export AWS_ENDPOINT_URL=http://localhost:4566
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
```

---

### Basic Usage
Compare live instances against terraform state file
```sh
./ec2diff --file ./examples/resources/terraform.tfstate 
```

Or, using Docker:

```sh
docker run --rm -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY \
   -v $(pwd):/data ec2diff \
   --file /data/examples/resources/terraform.tfstate 
```
---

Check on specific attributes:
```sh
./ec2diff --file ./examples/resources/terraform.tfstate --attrs="instance_type,tags"
```

To get a list of supported attributes run:
```sh
./ec2diff --list-attributes
```
---

You can see a list of available arguments by running help:
```sh 
./ec2diff -h
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

——

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
   - [`PaginatedLiveFetcher`](./pkg/livefetcher.go) interface for fetching instances from a live source (e.g. AWS).
   - [`Parser`](./pkg/parser.go) interface for parsing state files passed to the program to extract instance definitions.
   - [`DriftChecker`](./pkg/driftchecker.go) interface abstracts logic for comparing instances to detect differences/drifts.
   - [`ReportPrinter`](./pkg/reportprinter.go) interface abstracts logic for presenting/printing reports of drifts.
- [**registry**](./registry) registers available parsers. Associates provided file type to a parser for parsing.
- [main.go](./main.go) the program's entry point.

---

### 🔄 Execution Flow

- 🛠️ **Parse** – The specified file is parsed using a registered parser based on its type (`.tfstate` or `.json`). This extracts all EC2-related state resources into memory for comparison.
- 📥 **Fetch** – Live EC2 resources are retrieved from AWS using efficient pagination. Each page provides a batch of live instances for analysis.
- ⚖️ **Compare** – For every page of live instances, the drift checker runs concurrently to compare them against the parsed state. The result is a list of drift reports.
- 🧾 **Report** – All drift reports are collected and printed to standard output in a readable table format.

---

### 🔍 State Comparison Approach

- The tool compares live resources on AWS against the desired state (Terraform state file).
- ✅ Comparison is attribute-driven, allowing users to specify which fields to check—enabling flexibility and performance tuning.
- 🧩 A common state interface ([`pkg.Instance`](./pkg/instance.go)) supports easy extensibility for new attribute fields and file formats.

---

### ⚙️ Concurrency

- Drift detection is performed concurrently for each instance using a goroutine-based worker pool.
- This approach improves performance while minimizing resource consumption—ideal for large infrastructures.

---

### ❗ Error Handling

- Errors are surfaced with clear log messages and trigger immediate program termination to prevent partial or misleading results.

---

### 🚀 Scalability & Tolerance

- Efficient pagination during resource fetching, combined with a configurable worker pool for drift checks, allows the program to handle large-scale workloads reliably and with minimal overhead.

---

### ⚠️ Trade-offs & Limitations

- 🔍 **Attribute Comparison** – Deeply nested attributes may not fully compared automatically, as reflection-based traversal is avoided for performance reasons.
- 🧠 **State File Parsing**
  - All state resources are parsed and loaded into memory in a single pass, ensuring complete visibility during drift checks.
  - ⚠️ Performance may degrade with very large state files due to the non-streamed, in-memory approach.

---

### 🔧 Extensibility

- Modular design—segregated packages for AWS integration, Terraform parsing, and drift detection—makes it easy to extend support for new resource types or input formats.
- Using an explicit list of attributes allows seamless addition of new supported fields.

---

### 🛠️ Language & Libraries

- **Go** – Chosen for its rich concurrency model, static typing, and effectiveness for CLI development.
- **Flags** – Lightweight CLI parser for clean argument handling and help generation.
- **AWS SDK for Go v2** – Used for performant and modern AWS API interaction.
- **HCL** – Parses infrastructure definitions from Terraform configurations.
- **go-cmp** – Provides safer and more flexible attribute comparisons than `reflect.DeepEqual`.


## Future Improvements
* Export drift reports in formats such as JSON, HTML, or CSV