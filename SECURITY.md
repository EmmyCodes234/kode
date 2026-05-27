# Security Policy

## Supported Versions

Currently, only the latest release of Kode is actively supported with security updates.

| Version | Supported          |
| ------- | ------------------ |
| v1.15.x | :white_check_mark: |
| < v1.15 | :x:                |

## Reporting a Vulnerability

Security is the core thesis of Kode. If you discover a vulnerability in the verification gates, AST parsing, MCP server, or any other component, please report it immediately.

**Do not open a public issue.**

Instead, please email security@sicario.io with a detailed description of the vulnerability and steps to reproduce it. We will acknowledge your report within 24 hours and provide an estimated timeline for the fix.

## Zero-Exfiltration Guarantee
Kode is designed under a strict "Zero-Exfiltration" philosophy. Session data, codebase context, and user prompts are NEVER sent to Sicario Labs or any third-party analytics service. The only external connections made are direct HTTPS requests from your localhost to your explicitly configured LLM provider (e.g., OpenAI, Anthropic). 

If you find any deviation from this guarantee, it is considered a critical security vulnerability.
