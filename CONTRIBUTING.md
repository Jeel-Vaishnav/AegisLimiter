# Contributing to AegisLimiter

Thank you for your interest in contributing to **AegisLimiter**! We welcome community contributions, bug reports, algorithmic improvements, and feature proposals.

---

## 🛠️ Development Setup

### Prerequisites
- **Go 1.22+**
- **Docker & Docker Compose** (optional for Redis and Grafana telemetry)
- **Python 3.10+** (for benchmark harnesses)

### Setting Up Locally

1. Fork and clone the repository:
   ```bash
   git clone https://github.com/Jeel-Vaishnav/AegisLimiter.git
   cd AegisLimiter
   ```

2. Download Go dependencies:
   ```bash
   go mod download
   ```

3. Run the unit and concurrency test suite:
   ```bash
   go test -v -race ./tests/...
   ```

4. Run the microbenchmarks:
   ```bash
   go test -bench=. -benchmem ./tests/...
   ```

---

## 📋 Pull Request Process

1. **Create a Feature Branch**:
   ```bash
   git checkout -b feat/your-feature-name
   ```
2. **Commit Guidelines**:
   We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:
   - `feat:` New features or algorithms
   - `fix:` Bug fixes or race condition corrections
   - `docs:` Documentation improvements
   - `perf:` Performance optimizations
   - `test:` Adding or updating tests
   - `refactor:` Code restructuring without functional changes

3. **Verify Tests Pass**:
   Ensure all tests pass with the Go race detector enabled:
   ```bash
   go test -race ./tests/...
   ```

4. **Submit Your Pull Request**:
   Open a PR against the `main` branch with a clear description of the problem solved and benchmark metrics where applicable.
