# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based MQTT to command-line applications gateway (`mqtt2cmd`) that creates virtual MQTT switches from command-line applications. It allows exposing local CLI applications to home automation servers like Home Assistant via MQTT.

## Common Development Commands

### Building and Testing
```bash
# Build the application
go build -o mqtt2cmd .

# Run tests (if any exist)
go test ./...

# Run the application
./mqtt2cmd

# Build with version information (for releases)
go build -ldflags "-X main.version=1.0.0 -X main.commit=abc123 -X main.date=2023-10-01" .
```

### Dependencies

**Go versions**: Use https://go.dev/doc/devel/release for the latest stable Go versions.
```bash
# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

### Release Building (using GoReleaser)
```bash
# Install GoReleaser (if not already installed)
go install github.com/goreleaser/goreleaser@latest

# Build release artifacts
goreleaser build --snapshot

# Full release (requires proper Git tags and GitHub setup)
goreleaser release
```

### Docker
```bash
# Build Docker image
docker build -t mqtt2cmd .

# Run Docker container
docker run mqtt2cmd
```

## Architecture

### Core Components

1. **Entry Point** (`main.go`):
   - Simple main function that delegates to `cmd.Execute()`
   - Handles version information injection via build flags

2. **Command Layer** (`cmd/app.go`):
   - Orchestrates application startup and configuration
   - Initializes logging and MQTT client
   - Executes the main application loop with periodic refresh

3. **Configuration** (`internal/config/config.go`):
   - Handles configuration loading from multiple sources (config file, environment variables, command-line flags)
   - Supports YAML configuration files with platform-specific default locations
   - Uses Viper for configuration management with environment variable prefix `MQTT2CMD`

4. **Switch Controls** (`internal/controls/switches.go`):
   - Defines the `Switch` struct with commands for turning on/off and getting state
   - Executes shell commands and interprets exit codes (exit code 0 = ON, exit code 1 = OFF)
   - All commands are executed via `/bin/sh -c`

5. **MQTT Integration** (`internal/mqtt/mqtt.go`):
   - Handles MQTT client initialization and communication
   - Manages topic subscriptions and publications
   - Implements device availability and state publishing

### Configuration Structure

The application uses a YAML configuration file with the following structure:
```yaml
app-id: 'mqtt2cmd'  # MQTT topic prefix
mqtt:
  broker: "tcp://your-mqtt-server-address:1883"
  username: "optional-username"
  password: "optional-password"
switches:
  - name: "switch-name"
    turn_on: "/path/to/command/on"
    turn_off: "/path/to/command/off"
    get_state: "/path/to/command/state"
    refresh: "10m"  # refresh interval
log:
  path: "/path/to/logfile.log"
```

### MQTT Topic Structure

- Application availability: `{app-id}/available`
- Switch availability: `{app-id}/switches/{switch-name}/available`
- Switch state: `{app-id}/switches/{switch-name}`
- Switch control: `{app-id}/switches/{switch-name}/set`

### Key Dependencies

- **MQTT Client**: `github.com/eclipse/paho.mqtt.golang`
- **Configuration**: `github.com/spf13/viper` and `github.com/spf13/pflag`
- **Logging**: `go.uber.org/zap` with `gopkg.in/natefinch/lumberjack.v2` for log rotation

### Platform-Specific Configuration Locations

- **macOS**: `$HOME/Library/Application Support/mqtt2cmd/config.yaml`
- **Linux**: `$XDG_CONFIG_HOME/mqtt2cmd/config.yaml` or `$HOME/.config/mqtt2cmd/config.yaml`

### Build Configuration

- Uses GoReleaser for automated builds and releases
- Supports multi-platform builds (Linux/AMD64, Linux/ARM64, Darwin/AMD64, Darwin/ARM64)
- Creates Homebrew formula for macOS distribution
- Builds static binaries with CGO_ENABLED=0
- Injects version information via ldflags

### Application Flow

1. Load configuration from file, environment variables, and command-line flags
2. Initialize structured logging with file output and rotation
3. Connect to MQTT broker with configured credentials
4. Subscribe to control topics for all configured switches
5. Enter main loop: refresh switch states periodically (every 10 seconds)
6. Handle incoming MQTT messages to switch control topics
7. Execute corresponding commands and publish state changes

### Logging

- Uses Uber's Zap logger for high-performance structured logging
- Log rotation handled by Lumberjack
- MQTT client instrumentation is enabled for debugging communication