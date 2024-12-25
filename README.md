# THCloud CLI Tool (thctl)

A command-line tool for managing Filecoin storage providers, focusing on sector management and financial calculations.

## Features

- Filecoin sector management
  - Calculate sector penalties
  - Query vested funds
- Cloud Storage Operations
  - AWS S3 management
  - Aliyun OSS management
  - Tencent COS management
- More features coming soon...

## Installation

### Prerequisites
- Go 1.21 or higher
- Access to a Lotus node

### Install from source
```bash
# Clone the repository
git clone https://github.com/thcloudai/thctl.git

# Change to project directory
cd thctl

# Build and install
go install ./cmd/thctl
```

## Configuration

1. Create a configuration directory:
```bash
mkdir -p ~/.thctl
```

2. Create a configuration file (`~/.thctl/config.yaml`):
```yaml
lotus:
  api_address: "http://127.0.0.1:1234/rpc/v0"  # Your Lotus node API address
  auth_token: ""                                # Your Lotus API token

log:
  level: "info"                                # Log level (debug, info, warn, error)
  file: ""                                     # Log file path (empty for stdout)
```

3. Set your Lotus API token (alternatively, you can set it in the config file):
```bash
export THCTL_LOTUS_AUTH_TOKEN="your-lotus-token"
```

## Cloud Storage Configuration

### AWS S3
```bash
# Set credentials via environment variables
export THCTL_S3_ACCESS_KEY="your-access-key"
export THCTL_S3_SECRET_KEY="your-secret-key"
```

Or configure in `~/.thctl/config.yaml`:
```yaml
storage:
  s3:
    region: "us-west-2"
    access_key: "your-access-key"
    secret_key: "your-secret-key"
    endpoint: ""  # Optional, for S3-compatible services
    bucket_name: "default-bucket"  # Optional default bucket
```

### Aliyun OSS
```bash
# Set credentials via environment variables
export THCTL_OSS_ACCESS_KEY="your-access-key"
export THCTL_OSS_SECRET_KEY="your-secret-key"
```

Or configure in `~/.thctl/config.yaml`:
```yaml
storage:
  oss:
    region: "cn-hangzhou"
    access_key: "your-access-key"
    secret_key: "your-secret-key"
    endpoint: "oss-cn-hangzhou.aliyuncs.com"
    bucket_name: "default-bucket"
    internal: false  # Use internal endpoint
```

### Tencent COS
```bash
# Set credentials via environment variables
export THCTL_COS_ACCESS_KEY="your-access-key"
export THCTL_COS_SECRET_KEY="your-secret-key"
```

Or configure in `~/.thctl/config.yaml`:
```yaml
storage:
  cos:
    region: "ap-guangzhou"
    access_key: "your-access-key"
    secret_key: "your-secret-key"
    endpoint: ""  # Optional COS endpoint
    bucket_name: "default-bucket"
    app_id: "your-app-id"
```

## Authentication

There are two ways to authenticate with THCloud.AI:

### 1. Interactive Authentication (Recommended)

Run the following command to start the authentication flow:
```bash
thctl auth
```

This will:
1. Open your default browser
2. Direct you to the THCloud.AI login page
3. After successful login, save your credentials locally

The credentials are stored in `~/.config/thctl/th-credentials.json` (location may vary by OS).

### 2. API Key Authentication

You can authenticate using an API key in two ways:

a. Using the `--api-key` flag:
```bash
thctl --api-key=your-api-key [command]
```

b. Using environment variable:
```bash
export THC_API_KEY=your-api-key
thctl [command]
```

## Global Options

The following options are available for all commands:

```
  -o, --output      Set output format
                    [string] [choices: "json", "yaml", "table"] [default: "table"]
  --config-dir      Path to config directory
                    [string] [default: "~/.config/thctl"]
  --api-key         API key for authentication
                    [string] [default: ""]
  -v, --version     Show version number
                    [boolean]
  -h, --help        Show help
                    [boolean]
```

Example usage:
```bash
# Get output in JSON format
thctl fil sectors list --miner f01234 -o json

# Use custom config directory
thctl --config-dir=/path/to/config auth

# Show version
thctl -v

# Show help
thctl -h
```

## Usage

### Basic Commands

```bash
# Get help
thctl --help

# Get version
thctl version

# Get help for specific command
thctl fil sectors --help
thctl s3 --help
```

### Filecoin Commands

#### Calculate Sector Penalty
Calculate the penalty for a specific sector:
```bash
thctl fil sectors penalty --miner f01234 --sector 1
```

Options:
- `--miner`: Miner ID (required)
- `--sector`: Sector number (required)

Example output:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "penalty": "1000000000000000000"
  }
}
```

#### Query Vested Funds
Get the total vested funds for a miner:
```bash
thctl fil sectors vested --miner f01234
```

Options:
- `--miner`: Miner ID (required)

Example output:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "vested": "5000000000000000000"
  }
}
```

### Cloud Storage Commands

#### AWS S3 Operations
```bash
# List buckets or objects
thctl s3 ls [bucket]

# Upload files
thctl s3 upload [source] [bucket/key]

# Download files
thctl s3 download [bucket/key] [destination]

# Delete objects
thctl s3 rm [bucket/key]
```

#### Aliyun OSS Operations
```bash
# List buckets or objects
thctl oss ls [bucket]

# Upload files
thctl oss upload [source] [bucket/key]

# Download files
thctl oss download [bucket/key] [destination]

# Delete objects
thctl oss rm [bucket/key]
```

#### Tencent COS Operations
```bash
# List buckets or objects
thctl cos ls [bucket]

# Upload files
thctl cos upload [source] [bucket/key]

# Download files
thctl cos download [bucket/key] [destination]

# Delete objects
thctl cos rm [bucket/key]
```

## Project Structure

```
thctl/
├── cmd/                      # Command line interface
│   └── thctl/
│       └── commands/        # Command implementations
│           ├── fil/        # Filecoin commands
│           ├── s3/         # AWS S3 commands
│           ├── oss/        # Aliyun OSS commands
│           └── cos/        # Tencent COS commands
├── internal/                # Private application code
│   ├── fil/                # Filecoin business logic
│   └── lotus/              # Lotus API client
├── pkg/                    # Public packages
│   └── framework/          # Framework components
│       ├── config/         # Configuration management
│       └── logger/         # Logging utilities
├── configs/                # Configuration files
│   └── config.example.yaml # Example configuration
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── LICENSE               # Apache 2.0 license
└── README.md            # Project documentation
```

## Common Issues

### Cannot connect to Lotus node
1. Verify your Lotus node is running
2. Check the API address in your config file
3. Ensure your API token is correct
4. Check network connectivity to the Lotus node

### Permission denied
Make sure you have the correct permissions and API token for the requested operation.

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the Apache-2.0 license - see the [LICENSE](LICENSE) file for details.