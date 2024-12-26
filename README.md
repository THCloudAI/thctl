# THCloud CLI Tool (thctl)

THCloud CLI Tool is a command-line interface for managing cloud storage and Filecoin nodes. It provides a unified interface for interacting with various cloud storage providers and Filecoin networks.

## Features

- **Cloud Storage Management**
  - Support for multiple cloud storage providers:
    - Aliyun OSS
    - AWS S3
    - Tencent COS
    - Huawei OBS
    - MinIO
  - File upload and download
  - Bucket management
  - Object listing and manipulation

- **Filecoin Node Management**
  - Query miner information
  - Sector management
  - Power statistics
  - Node status monitoring

## Installation

### Prerequisites

- Go 1.19 or higher
- Git

### Building from Source

```bash
# Clone the repository
git clone https://github.com/THCloudAI/thctl.git
cd thctl

# Build the binary
go build -o thctl ./cmd/thctl

# Optional: Move to system path
sudo mv thctl /usr/local/bin/
```

## Configuration

The THCloud CLI tool uses environment variables for configuration. You can set these variables in three ways:

1. Using a `.thctl.env` file:
   ```bash
   # Copy the example configuration
   cp .thctl.env.example .thctl.env
   
   # Edit the configuration with your values
   vim .thctl.env
   ```

2. Setting environment variables directly:
   ```bash
   export LOTUS_API_URL=http://127.0.0.1:1234/rpc/v0
   export LOTUS_API_TOKEN=your_token_here
   ```

3. Using command-line flags (these override environment variables):
   ```bash
   thctl fil --miner f0xxxx --api-url http://127.0.0.1:1234/rpc/v0
   ```

For detailed configuration options, see [.thctl.env.example](.thctl.env.example).

## Usage Examples

### Filecoin Commands

1. Query Miner Information:
   ```bash
   thctl fil --miner f0xxxx
   ```
   This will display detailed information about the specified miner, including:
   - Basic miner information (owner, worker, control addresses)
   - Current power statistics
   - Sector size and configuration

2. List Sectors:
   ```bash
   thctl fil sectors list --miner f0xxxx
   ```

3. Check Sector Status:
   ```bash
   thctl fil sectors status --miner f0xxxx --sector-id 1
   ```

### Cloud Storage Commands

1. Upload a File:
   ```bash
   thctl oss cp ./local/file.txt oss://bucket/remote/path/
   ```

2. Download a File:
   ```bash
   thctl oss cp oss://bucket/remote/file.txt ./local/path/
   ```

3. List Objects:
   ```bash
   thctl oss ls oss://bucket/prefix/
   ```

## Output Formats

The tool supports multiple output formats:

```bash
# JSON output (default)
thctl fil --miner f0xxxx -o json

# YAML output
thctl fil --miner f0xxxx -o yaml

# Table output
thctl fil --miner f0xxxx -o table
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Security

The `.thctl.env` file may contain sensitive information and should never be committed to version control. The file is already included in `.gitignore` to prevent accidental commits.

## Support

For support, please open an issue in the GitHub repository or contact the THCloud.AI team.