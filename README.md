# jet-access
super fast, simple, and secure ssh access portal

## Development

### Setup

1. Install Go 1.24.1 or later:
   - Download from [golang.org/dl](https://golang.org/dl/)
   - Verify installation with `go version`

2. Clone the repository:
   ```bash
   git clone https://github.com/Stone-IT-Cloud/jet-access.git
   cd jet-access
   ```

3. Install required tools:
   ```bash
   # Install golangci-lint
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

   # Install pre-commit hooks
   pip install pre-commit
   pre-commit install
   ```

### Configuration

1. Environment variables:
   - Create a `.env` file in the project root (see `.env.example` for required variables)
   - Alternatively, set environment variables in your shell

2. Application configuration:
   - Main configuration files are stored in the `configs/` directory
   - Modify `configs/config.yaml` for your development environment

### Development

1. Initialize the project:
   ```bash
   make download  # Download dependencies
   make tidy      # Ensure go.mod is tidy
   ```

2. Make code changes:
   - Follow the project structure:
     - `cmd/`: Main applications
     - `internal/`: Private packages
     - `pkg/`: Public packages
     - `tests/`: Integration tests

3. Run linting and tests:
   ```bash
   make fmt       # Format code
   make lint      # Run linter
   make test      # Run tests
   ```

### Building the software

1. Local development build:
   ```bash
   make build
   ```

2. The binary will be available at `build/bin/jet-access`

3. Run the built binary:
   ```bash
   ./build/bin/jet-access
   ```

4. Alternatively, run directly without building:
   ```bash
   make run
   ```

### Debugging

1. VS Code debugging:
   - Use the provided launch configurations in `.vscode/launch.json`
   - Set breakpoints in your code and press F5 to start debugging

2. Command-line debugging with Delve:
   ```bash
   # Install Delve
   go install github.com/go-delve/delve/cmd/dlv@latest

   # Run debugger
   dlv debug ./cmd/jet-access
   ```

3. Analyzing logs:
   - Logs are written to stdout/stderr by default
   - Increase verbosity using the appropriate command-line flags

4. Profiling:
   ```bash
   # Enable Go profiling
   go tool pprof [profile_file]
   ```
