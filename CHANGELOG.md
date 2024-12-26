# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-12-25

### Added
- Initial release of THCloud CLI Tool
- Authentication command with browser-based flow
- Filecoin sector management commands
- Cloud storage support for AWS S3, Aliyun OSS, and Tencent COS
- Global options for output format and configuration
- Comprehensive documentation and examples
- Logging system with file rotation
- Configuration management with environment variables support

### Security
- Secure credential storage
- API key support via environment variables
- Safe handling of sensitive information

## [Unreleased]

### Planned
- Additional cloud storage providers
- Enhanced error handling and recovery
- More detailed logging and debugging options
- Performance improvements
- Additional test coverage
=======
## [v0.1.0] - 2024-12-26

### Added
- Initial release of THCloud CLI tool
- Filecoin sector management commands:
  - `sectors list`: List all sectors for a miner
  - `sectors status`: Get sector status
  - `sectors info`: Get detailed sector information
  - `sectors penalty`: Query sector penalties
  - `sectors vested`: Check vested funds
- Support for multiple storage providers:
  - Filecoin
  - OSS
  - COS
  - S3
- Configurable output formats:
  - JSON
  - YAML
  - Table
- Authentication system
- Configuration management
- Logging system
- Documentation:
  - README.md
  - Command documentation
  - Configuration examples

### Changed
- Optimized build and installation process
- Improved error handling in Lotus client
- Enhanced configuration file structure

[v0.1.0]: https://github.com/THCloudAI/thctl/releases/tag/v0.1.0
