# The THCloud AI Command Line tools

## Description
The THCloud AI CLI is a command-line interface that lets you manage data storage,compute resource,etc.directly from the terminal.

## Project Structure
```tree
thctl/                 # Project root directory
├── api/                     # API protocol definitions
│   ├── http/               # HTTP API definitions
│   ├── grpc/               # gRPC protocol files (.proto)
│   └── swagger/            # Swagger/OpenAPI documentation
├── cmd/                     # Main applications entries
│   └── server/             # Server application entry
│       └── main.go         # Main program entry point
├── configs/                 # Configuration files directory
│   ├── config.yaml         # Application configuration file
│   └── config.example.yaml # Example configuration file
├── internal/               # Private application code
│   ├── app/               # Application core logic
│   │   └── service/       # Business service implementations
│   ├── pkg/               # Private packages
│   │   ├── db/           # Database related code
│   │   ├── middleware/   # Middleware components
│   │   └── utils/        # Utility functions
│   └── server/           # Server implementations
│       └── http/         # HTTP server
├── pkg/                    # Libraries that can be used by external applications
│   ├── logger/            # Logging package
│   └── errors/            # Error handling package
├── scripts/                # Scripts directory
│   ├── build.sh           # Build scripts
│   └── deploy.sh          # Deployment scripts
├── test/                   # Testing directory
│   ├── integration/       # Integration tests
│   └── mock/             # Test mock data
├── web/                    # Web related files (if applicable)
│   ├── static/           # Static files
│   └── template/         # Template files
├── .gitignore             # Git ignore file
├── Dockerfile             # Docker build file
├── go.mod                 # Go module definition
├── go.sum                 # Go module dependency checksums
├── Makefile               # Project management commands
└── README.md              # Project documentation
```

## Requirements
- Go 1.21 or higher
- Other dependencies...

## Getting Started

### Installation
```bash
# Clone the repository
git clone https://github.com/thcloudai/thctl.git

# Change to project directory
cd thctl

# Install dependencies
go mod download
```

### Configuration
Describe how to configure your project...

### Running the Application
```bash
# Run the application
go run cmd/server/main.go
```

## Contributing
Instructions for how to contribute to the project...

## License
This project is licensed under the Apache-2.0 license - see the [LICENSE](https://github.com/THCloudAI/thctl?tab=Apache-2.0-1-ov-file) file for details.