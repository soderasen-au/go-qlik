# go-qlik

A comprehensive Go SDK for interacting with Qlik Sense (on-premises) and Qlik Cloud Services (QCS).

## Features

- **Engine API**: WebSocket connections to Qlik Engine for app manipulation via enigma-go
- **QRS (Repository Service)**: REST API for on-premises Qlik Sense administration
- **QPS (Proxy Service)**: REST API for proxy operations
- **QCS**: REST API for Qlik Cloud Services
- **NPrinting**: Integration with Qlik NPrinting
- **Reporting**: Generate reports from Qlik apps in multiple formats (PDF, Excel, CSV, TSV)
- **Script Suite (ss)**: Task automation framework for executing sequences of Qlik operations

## Command-line Tools

### pdfprinter

Standalone tool for generating reports from Qlik apps with extensive customization options.

**Features:**
- Multiple output formats: PDF, Excel (XLSX), CSV, TSV
- PDF orientation support (portrait/landscape)
- Bookmark support for pre-filtered data
- Customizable output paths and report names
- Current selections display

**Usage:**
```bash
# Generate PDF report (default)
./bin/pdfprinter

# Generate Excel report
./bin/pdfprinter -format xlsx

# Custom configuration
./bin/pdfprinter \
  -app-id "your-app-id" \
  -bm-id "your-bookmark-id" \
  -format pdf \
  -orientation landscape \
  -name "Q4SalesReport" \
  -output-folder "./reports"

# View all options
./bin/pdfprinter -h
```

### jwt

JWT encoder/decoder utility for Qlik authentication.

## Installation

### Prerequisites

- Go 1.21 or later
- Access to a Qlik Sense or Qlik Cloud instance

### Clone and Build

```bash
git clone https://github.com/soderasen-au/go-qlik.git
cd go-qlik

# Build all binaries
make build

# Or build specific tools
make build-pdfprinter
make build-jwt
```

Binaries will be available in the `bin/` directory.

## Quick Start

### Connecting to Qlik Engine

```go
package main

import (
    "github.com/soderasen-au/go-qlik/qlik/engine"
    "github.com/soderasen-au/go-common/crypto"
)

func main() {
    cfg := engine.Config{
        EngineURI:     "wss://your-qlik-server:4747",
        AppID:         "your-app-id",
        UserName:      "username",
        UserDirectory: "directory",
        AuthMode:      engine.AUTH_MODE_CERT,
        ServerType:    engine.ST_ON_PREM,
        Certs: crypto.Certificates{
            ClientFile:    "/path/to/client.pem",
            ClientkeyFile: "/path/to/client_key.pem",
            CAFile:        "/path/to/root.pem",
        },
    }

    conn, err := engine.NewConn(cfg)
    if err != nil {
        // Handle error
    }
    defer conn.Disconnect()

    doc, err := conn.Global.OpenDoc(engine.ConnCtx, cfg.AppID, "", "", "", false)
    if err != nil {
        // Handle error
    }
    defer doc.DisconnectFromServer()

    // Work with your app...
}
```

### Generating Reports

```go
package main

import (
    "github.com/soderasen-au/go-qlik/report"
    "github.com/soderasen-au/go-common/util"
)

func main() {
    // Assume doc is already opened

    printer := report.NewBuiltInReportPrinter()
    r := report.Report{
        Name:         util.Ptr("SalesReport"),
        AppId:        "your-app-id",
        Doc:          doc,
        Target:       report.TARGET_OBJECTS,
        TargetIDs:    []string{"object-id"},
        OutputFormat: util.Ptr(report.REPORT_FORMAT_PDF),
        OutputFolder: util.Ptr("./reports"),
        OutputPDFOrientation: util.Ptr(report.PDF_ORIENTATION_PORTRAIT),
        Logger:       logger,
    }

    res := printer.Print(r)
    if res != nil {
        // Handle error
    }
}
```

## Authentication

The SDK supports three authentication modes:

### 1. Certificate Authentication (On-Premises)

```go
cfg := engine.Config{
    AuthMode: engine.AUTH_MODE_CERT,
    Certs: crypto.Certificates{
        ClientFile:    "/path/to/client.pem",
        ClientkeyFile: "/path/to/client_key.pem",
        CAFile:        "/path/to/root.pem",
    },
}
```

### 2. JWT Authentication (On-Premises)

```go
cfg := engine.Config{
    AuthMode: engine.AUTH_MODE_JWT,
    Certs: crypto.Certificates{
        ClientFile:    "/path/to/private_key.pem",
        ClientkeyFile: "/path/to/public_key.pem",
    },
}
```

### 3. Desktop Mode (Local Development)

```go
cfg := engine.Config{
    AuthMode: engine.AUTH_MODE_DESKTOP,
}
```

## Development

### Using the Makefile

```bash
# Build all binaries
make build

# Run all tests
make test

# Run specific package tests
make test-engine
make test-qrs
make test-report

# Generate test reports
make test-pdf

# Code quality
make fmt vet lint

# Clean build artifacts
make clean

# Show all available commands
make help
```

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./qlik/engine
go test ./qlik/managed/qrs
go test ./report

# With coverage
make test-coverage
```

### Project Structure

```
qlik/
├── client/         # High-level managed client
├── config/         # Configuration structs
├── engine/         # Qlik Engine JSON API and WebSocket connections
├── managed/        # On-premises Qlik Sense APIs (QRS, QPS)
├── qcs/            # Qlik Cloud Services APIs
└── rac/            # REST API Client base

report/
├── report.go       # Report generation framework
├── excel.go        # Excel output
├── csv.go          # CSV/TSV output
├── pdf.go          # PDF output
└── color.go        # Color handling

ss/                 # Script Suite - task automation framework
├── env.go          # Execution environment
├── script.go       # Script definitions
└── task_*.go       # Task implementations

cmd/
└── jwt/            # JWT encoder/decoder CLI

test/
└── pdf/            # PDF printer CLI tool
```

## Configuration

Configuration can be provided via YAML files. Example:

```yaml
sense:
  engine:
    engine_uri: wss://your-server:4747/app
    server_type: on_prem
    auth_mode: cert
    user_id: username
    user_directory: directory
    certs:
      client: /path/to/client.pem
      client_key: /path/to/client_key.pem
      root_ca: /path/to/root.pem
  qrs:
    base_url: https://your-server:4242
    auth:
      method: cert
      user: username
      certs:
        client: /path/to/client.pem
        client_key: /path/to/client_key.pem
        root_ca: /path/to/root.pem
```

## Dependencies

This project uses a local dependency on `github.com/soderasen-au/go-common`. Ensure it exists at `../go-common` relative to this repository.

To update dependencies:

```bash
make deps
```

## Error Handling

The SDK uses `*util.Result` from go-common for error handling instead of standard Go errors:

```go
res := someOperation()
if res != nil {
    return res.With("additional context")
}
```

## Logging

Structured logging via `github.com/rs/zerolog`:

```go
logger, _ := loggers.GetLogger("app.log")
client := engine.NewHttpClient(cfg, logger)
```

## Report Formats

### Supported Formats

- **PDF**: Portrait or landscape orientation, with color support
- **Excel (XLSX)**: Multi-sheet support, cell formatting, colors
- **CSV**: Comma-separated values
- **TSV**: Tab-separated values

### Report Targets

- `TARGET_OBJECTS`: Specific Qlik objects by ID
- `TARGET_SHEET`: Entire sheet export

### Report Drivers

- `DRIVER_SENSE`: Use Qlik's native hypercubes
- `DRIVER_BUILT_IN`: Custom report generation logic

## Examples

See `test/pdf/main.go` for a complete example of:
- Connecting to Qlik Engine
- Opening an app
- Applying bookmarks
- Generating reports in multiple formats

## Contributing

Contributions are welcome! Please ensure:

1. Code is formatted with `make fmt`
2. All tests pass with `make test`
3. Linting passes with `make vet`

## License

[Specify your license here]

## Acknowledgments

- Built on [enigma-go](https://github.com/qlik-oss/enigma-go) for Qlik Engine communication
- Uses [excelize](https://github.com/qxnw/excelize) for Excel generation
- Uses [gofpdf](https://github.com/jung-kurt/gofpdf) for PDF generation
- Uses [zerolog](https://github.com/rs/zerolog) for structured logging

## Support

For issues, questions, or contributions, please open an issue on GitHub.
