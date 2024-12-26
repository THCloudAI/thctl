# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in THCloud CLI Tool, please report it by emailing our security team. We will respond to your report within 24 hours.

## Security Best Practices

### Configuration Management
1. Never commit sensitive information to the repository
2. Use environment variables for sensitive data:
   ```bash
   export FULLNODE_API_INFO="your-token:/ip4/127.0.0.1/tcp/1234/http"
   export THC_API_KEY="your-api-key"
   ```

3. Use configuration files only for non-sensitive settings
4. Keep production configuration files separate and never commit them

### API Keys and Tokens
1. Rotate API keys and tokens regularly
2. Use different API keys for development and production
3. Set appropriate permissions and access levels
4. Never share API keys in logs or error messages

### Development Guidelines
1. Always run security checks before committing:
   ```bash
   make lint
   make test
   ```

2. Use the pre-commit hook (automatically installed)
3. Keep dependencies up to date
4. Follow secure coding practices:
   - Input validation
   - Error handling without exposing sensitive info
   - Proper logging (no sensitive data)

### Pre-commit Checks
The repository includes a pre-commit hook that checks for:
1. Sensitive information in code
2. Large files (>10MB)
3. Code quality (golangci-lint)
4. Security issues (gitleaks)

### Installation Security
1. Always verify checksums of downloaded binaries
2. Use HTTPS for downloading
3. Set appropriate file permissions
4. Keep your system and Go installation up to date

### Runtime Security
1. Run with minimum required permissions
2. Use secure network connections (HTTPS/TLS)
3. Validate all inputs
4. Handle errors securely

## Security Tools

### Required Tools
- [gitleaks](https://github.com/zricethezav/gitleaks) - Scan for secrets
- [golangci-lint](https://golangci-lint.run/) - Code quality and security
- [gosec](https://github.com/securego/gosec) - Go security checker

### Installation
```bash
# Install gitleaks
brew install gitleaks

# Install golangci-lint
brew install golangci-lint

# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

## Secure Development Workflow

1. **Before Starting Development**
   ```bash
   git pull
   go mod tidy
   make deps
   ```

2. **During Development**
   - Use environment variables for secrets
   - Follow secure coding guidelines
   - Run tests frequently

3. **Before Committing**
   ```bash
   make lint
   make test
   # Pre-commit hook will run automatically
   ```

4. **Before Pushing**
   ```bash
   # Run full security scan
   gitleaks detect
   gosec ./...
   ```

## Version Control Security

1. **Protected Branches**
   - `main` branch is protected
   - Requires pull request reviews
   - Requires passing CI checks

2. **Commit Signing**
   - Enable GPG signing of commits
   - Verify commit signatures

3. **Branch Protection Rules**
   - No force pushes
   - Required status checks
   - Required security reviews

## Dependency Management

1. **Regular Updates**
   ```bash
   go list -u -m all
   go get -u ./...
   go mod tidy
   ```

2. **Vulnerability Scanning**
   - Regular dependency audits
   - Automated security updates
   - Version pinning for stability

## Production Deployment

1. **Binary Verification**
   - Check SHA256 checksums
   - Verify binary signatures
   - Use official releases only

2. **Configuration**
   - Use environment-specific configs
   - Validate all settings
   - Monitor for misconfigurations

3. **Monitoring**
   - Enable error tracking
   - Monitor API usage
   - Set up alerts for suspicious activity
