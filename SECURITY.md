# Security Policy

## Supported Versions

Security fixes are applied to the latest release. Older versions are not backported unless the severity warrants it.

| Version | Support |
|---|---|
| Latest release | Security fixes applied |
| Older releases | Best effort |

## Reporting a Vulnerability

**Do not report security vulnerabilities in public GitHub Issues.**

Please report security vulnerabilities privately using [GitHub Security Advisories](https://github.com/taqu/meiliscope/security/advisories/new). This allows the maintainer to review and address the issue before it is publicly disclosed.

Include in your report:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Any suggested mitigations

You will receive a response acknowledging the report within 7 days. The expected timeline for a fix depends on severity.

## Security Model

MeiliScope is a local desktop application. It does not expose a network server or accept connections from external hosts.

For the full threat model, security properties, and known limitations, see [docs/architecture/security.md](docs/architecture/security.md).

## Credential Handling

API keys are stored in the OS keychain (Windows Credential Manager / Linux Secret Service) and are never written to configuration files or log files.

## Scope

The following are within scope for security reports:

- Credential leakage (API keys visible outside the OS keychain)
- Unauthorized document mutations
- Cross-site-scripting via document content rendered in the application
- Redirect credential forwarding
- Insecure default configurations

The following are outside scope:

- Meilisearch server-side vulnerabilities
- Operating system vulnerabilities
- Vulnerabilities in Meilisearch itself
- Attacks requiring full OS-level compromise of the user's machine
