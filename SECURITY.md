# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability, please follow these steps:

### Do NOT

- Open a public GitHub issue
- Discuss the vulnerability publicly before it's fixed
- Exploit the vulnerability beyond what's necessary to demonstrate it

### Do

1. **Email**: Send details to security@example.com (replace with actual contact)
2. **Include**:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### What to Expect

- **Acknowledgment**: Within 48 hours
- **Initial Assessment**: Within 7 days
- **Resolution Timeline**: Depends on severity, typically 30-90 days
- **Credit**: We'll credit you in the release notes (unless you prefer anonymity)

## Security Considerations

### Device Authentication

- The CLI stores device credentials in the config file
- Config files should have restricted permissions (0600)
- Use environment variables for sensitive data in CI/CD

### Cloud Credentials

- OAuth tokens are stored securely in the config
- Tokens are refreshed automatically
- Use `shelly cloud logout` to clear credentials

### Plugin Security

- Only install plugins from trusted sources
- Plugins execute with the same permissions as the CLI
- Review plugin source code before installation

## Best Practices

1. **Keep Updated**: Use the latest version of the CLI
2. **Secure Config**: Ensure config files have proper permissions
3. **Network Security**: Use device authentication when possible
4. **Audit Plugins**: Review installed plugins regularly
