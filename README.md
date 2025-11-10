# AccelByte Extend Challenge Demo App

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

**CLI and TUI tool for testing the Challenge Suite locally.**

This demo app provides both command-line and terminal UI modes for interacting with the Challenge Service and Event Handler during local development and testing.

---

## Features

✅ **CLI Mode** - Command-line interface for scripting and automation
✅ **TUI Mode** - Interactive terminal UI for visual testing
✅ **Challenge Management** - List challenges, view progress, claim rewards
✅ **Event Triggering** - Simulate IAM login and Statistic update events
✅ **Mock Authentication** - No AGS credentials needed for local testing
✅ **Real AGS Support** - Can use real AGS credentials for integration testing

---

## Quick Start

### Installation

```bash
# Clone repository
git clone https://github.com/AccelByte/extend-challenge-demo-app.git
cd extend-challenge-demo-app

# Build
go build -o challenge-demo main.go

# Or run directly
go run main.go
```

### Basic Usage

**List all challenges**:
```bash
./challenge-demo challenges list
```

**Get specific challenge**:
```bash
./challenge-demo challenges get daily-quests
```

**Trigger login event**:
```bash
./challenge-demo events trigger login
```

**Trigger statistic event**:
```bash
./challenge-demo events trigger stat --stat-code=match_wins --value=5
```

**Claim reward**:
```bash
./challenge-demo challenges claim daily-quests daily-login
```

**Launch TUI mode**:
```bash
./challenge-demo
# Or explicitly:
./challenge-demo tui
```

---

## Configuration

Create `.env` file (or set environment variables):

```bash
# Challenge Service
CHALLENGE_SERVICE_URL=http://localhost:8000
CHALLENGE_SERVICE_GRPC=localhost:6565

# Event Handler
EVENT_HANDLER_GRPC=localhost:6566

# Mock User (for local testing)
DEMO_USER_ID=test-user-123
DEMO_NAMESPACE=local

# Real AGS (optional, for integration testing)
AB_BASE_URL=https://your-environment.accelbyte.io
AB_CLIENT_ID=your-client-id
AB_CLIENT_SECRET=your-client-secret
AB_NAMESPACE=your-namespace
AUTH_MODE=mock  # or 'real' for AGS authentication
```

---

## CLI Commands

### Challenge Commands

```bash
# List all challenges
challenge-demo challenges list

# Get specific challenge by ID
challenge-demo challenges get <challenge-id>

# Claim reward for completed goal
challenge-demo challenges claim <challenge-id> <goal-id>
```

### Event Commands

```bash
# Trigger login event
challenge-demo events trigger login

# Trigger statistic update event
challenge-demo events trigger stat --stat-code=<code> --value=<number>

# Trigger multiple events
challenge-demo events trigger login --count=10
```

### Utility Commands

```bash
# Check service health
challenge-demo health

# Show configuration
challenge-demo config

# Show version
challenge-demo version
```

---

## TUI Mode

Launch interactive terminal UI:

```bash
./challenge-demo
```

**Controls**:
- `↑/↓` or `j/k` - Navigate lists
- `Enter` - Select item
- `c` - View challenges
- `e` - Trigger events
- `r` - Refresh data
- `q` or `Esc` - Quit/Back
- `?` - Help

**Screens**:
1. **Main Screen** - Overview of challenges and progress
2. **Challenge List** - Browse all challenges
3. **Challenge Detail** - View goals and claim rewards
4. **Event Trigger** - Simulate events interactively
5. **Settings** - Configure connection and auth

---

## Development

### Project Structure

```
extend-challenge-demo-app/
├── cmd/
│   ├── cli/                       # CLI commands
│   └── tui/                       # TUI screens
├── internal/
│   ├── client/                    # API clients
│   └── config/                    # Configuration
├── main.go                        # Entry point
├── go.mod
└── README.md
```

### Building

```bash
# Build for current platform
go build -o challenge-demo main.go

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o challenge-demo-linux main.go

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o challenge-demo-mac main.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o challenge-demo.exe main.go
```

### Testing

```bash
# Run tests
go test ./...

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Documentation

- **Suite Documentation**: https://github.com/AccelByte/extend-challenge-suite
- **Architecture Spec**: https://github.com/AccelByte/extend-challenge-suite/blob/master/docs/TECH_SPEC_M1.md
- **Demo App Design**: https://github.com/AccelByte/extend-challenge-suite/blob/master/docs/demo-app/

---

## Contributing

See [Suite repo - CONTRIBUTING.md](https://github.com/AccelByte/extend-challenge-suite/blob/master/CONTRIBUTING.md)

---

## License

[Apache 2.0 License](LICENSE)

---

## Links

- **Suite Repo**: https://github.com/AccelByte/extend-challenge-suite
- **Common Library**: https://github.com/AccelByte/extend-challenge-common
- **Backend Service**: https://github.com/AccelByte/extend-challenge-service
- **Event Handler**: https://github.com/AccelByte/extend-challenge-event-handler
