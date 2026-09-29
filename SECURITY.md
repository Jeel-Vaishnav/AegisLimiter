# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of **AegisLimiter** seriously. If you discover a security vulnerability or potential threat in this repository, please report it responsibly.

### How to Report

1. **Do NOT open a public GitHub issue** for sensitive vulnerabilities.
2. Send an advisory report via GitHub's [Security Advisory tab](https://github.com/Jeel-Vaishnav/AegisLimiter/security/advisories/new) or email `security@aegislimiter.dev`.
3. Include the following details:
   - Type of issue (e.g. race condition, unauthorized quota bypass, denial of service).
   - Component affected (e.g. `scripts/lua/token_bucket.lua`, `pkg/grpc/server.go`).
   - Step-by-step reproduction instructions or a minimal proof-of-concept.
   - Potential impact and suggested mitigations.

### Response Timeline

- **Acknowledgment**: Within 48 hours of initial report.
- **Assessment & Triage**: Within 5 business days.
- **Fix & Public Disclosure**: Coordinated release with a CVE/GHSA advisory once a patch is verified.

## Security Architecture & Best Practices

AegisLimiter is designed with security-by-default principles:
- **No Distributed Locks**: All state updates execute in atomic Redis Lua scripts to eliminate race-condition quota tampering.
- **Fail-Safe Defaults**: If Redis drops or becomes unresponsive, the service gracefully degrades to local memory boundaries rather than failing open.
- **Memory Bound Keys**: Dynamic TTLs on all Redis keys prevent memory starvation and malicious resource exhaustion attacks.
