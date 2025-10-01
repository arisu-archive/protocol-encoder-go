# Protocol Encoder

A high-performance protocol encoder built with Go and Unicorn Engine for encoding protocol data using ARM64 binary functions in a controlled emulation environment.

## Features

- **Protocol Encoding**: Encode protocol data using ARM64 binary functions
- **CRC32 Integration**: Built-in CRC32 checksum support for protocol validation
- **ARM64 Emulation**: Execute ARM64 encoding functions from binary files
- **High-Performance Pool**: Concurrent execution with emulator pooling for massive throughput
- **CLI Interface**: Command-line interface with JSON output support
- **Configuration Management**: YAML configuration files and environment variables
- **Structured Logging**: JSON-formatted logs with configurable verbosity
- **Performance Optimized**: Built in Go for high performance and low latency

## Quick Start

### Prerequisites

- Go 1.21 or later
- Unicorn Engine development libraries
- Your target binary file (`libil2cpp.so`)
- Function offset address

### Installation

```bash
# Clone the repository
git clone git@github.com:arisu-archive/protocol-encoder-go.git
cd protocol-encoder-go

# Install dependencies
make install

# Build the application
make build
```

### Basic Usage

```bash
# Encode protocol with default parameters
./bin/protocol-encoder-go

# Encode specific protocol with CRC32
./bin/protocol-encoder-go --protocol 0xDEADBEEF --crc32 0x12345678

# Output results in JSON format
./bin/protocol-encoder-go --json --protocol 0x12345678

# Use custom configuration file
./bin/protocol-encoder-go --config custom-config.yaml --protocol 0xABCDEF
```

## Configuration

### Command Line Options

```bash
Flags:
      --binary string       path to binary file (default "libil2cpp.so")
      --config string       config file (default is ./config.yaml)
      --crc32 uint64        CRC32 checksum for protocol validation
      --json                output results in JSON format
      --offset string       function offset (hex or decimal) (default "0x6268754")
      --pool-size int       emulator pool size for high throughput (1 = single emulator) (default 1)
      --protocol uint64     protocol value to encode (default 0xDEADBEEF)
  -v, --verbose             verbose output
```

### Configuration File

Create a `config.yaml` file:

```yaml
# Protocol Encoder Configuration
binary: "libil2cpp.so"
offset: 0x6268754

# Performance settings
pool_size: 1             # Emulator pool size (1 = single emulator, >1 = pool mode)
```

## Protocol Encoding

The encoder supports **flexible protocol encoding** through ARM64 function emulation:

### Examples:

```bash
# Encode protocol with CRC32 checksum
./bin/protocol-encoder-go --protocol 0xDEADBEEF --crc32 0x12345678

# Encode with custom offset
./bin/protocol-encoder-go --offset 0x6268754 --protocol 0xABCDEF

# High-throughput encoding with pool
./bin/protocol-encoder-go --pool-size 4 --protocol 0x12345678
```

### Parameter Mapping:
- **X0**: Protocol value to encode
- **X1**: CRC32 checksum for validation
- **Return Value**: Encoded protocol result

## Architecture

```
├── cmd/                 # Application entry point
│   └── cli/            # CLI interface and configuration
│       └── main.go     # Main application entry point
├── internal/           # Internal packages
│   └── emulator/       # Core emulation logic
│       ├── config.go   # Emulator configuration
│       ├── emulator.go # Main emulator implementation
│       ├── errors.go   # Error definitions
│       ├── emulator_test.go # Unit tests
│       └── benchmark_test.go # Performance benchmarks
├── pkg/                # Public packages
│   └── encoder/        # Protocol encoder package
│       ├── config.go   # Encoder configuration
│       ├── engine.go   # Encoder implementation
│       └── runner.go   # Emulator runner interface
├── config.yaml         # Default configuration
├── Makefile           # Build automation
└── README.md          # Documentation
```

### Key Components

- **Encoder**: High-level protocol encoding interface
- **Emulator**: Core ARM64 emulation engine with memory management
- **Config**: Type-safe configuration with validation
- **Error Handling**: Custom error types with context
- **CLI**: Cobra-based command-line interface
- **Logging**: Structured logging with configurable levels

## Development

### Available Make Targets

```bash
make help           # Show available commands
make build          # Build the application
make test           # Run tests
make test-coverage  # Run tests with coverage report
make lint           # Run linter
make fmt            # Format code
make clean          # Clean build artifacts
make dev-setup      # Set up development environment
make release        # Full release build
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test -v ./internal/emulator -run TestNew
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Full quality check
make release
```

## API Reference

### Encoder Interface

```go
// Create new encoder instance
encoder, err := encoder.NewEncoder(&encoder.Config{
    Binary:   "libil2cpp.so",
    Offset:   0x6268754,
    PoolSize: 1,
    Logger:   logger,
})

// Encode protocol with CRC32
result, err := encoder.Encode(protocol, crc32)

// Clean up resources
err = encoder.Close()
```

### Configuration

```go
// Encoder configuration
config := &encoder.Config{
    Binary:   "libil2cpp.so",
    Offset:   0x6268754,
    PoolSize: 1,
    Logger:   logger,
}

// Validate configuration
err := config.Validate()
```

### Encoding Result

```go
type InvokeResponse struct {
    ReturnValue     uint64        `json:"return_value"`
    FunctionAddress uint64        `json:"function_address"`
    ExecutionTime   time.Duration `json:"execution_time"`
    Success         bool          `json:"success"`
    Error           string        `json:"error,omitempty"`
}
```

## Performance

- **Memory Efficient**: Minimal memory allocation during protocol encoding
- **Fast Startup**: Optimized initialization sequence for quick encoding
- **Concurrent Safe**: Thread-safe design for parallel protocol encoding
- **Resource Management**: Automatic cleanup and leak prevention
- **High Throughput**: Emulator pooling for massive protocol encoding throughput

## Troubleshooting

### Common Issues

1. **Unicorn Engine Not Found**
   ```bash
   # Install Unicorn Engine development libraries
   # Ubuntu/Debian:
   sudo apt-get install libunicorn-dev
   
   # macOS:
   brew install unicorn
   ```

2. **Binary File Not Found**
   ```bash
   # Ensure your binary file exists
   ls -la libil2cpp.so
   
   # Or specify custom path
   ./bin/protocol-encoder-go --binary /path/to/your/binary.so
   ```

### Debug Mode

```bash
# Enable verbose logging
./bin/protocol-encoder-go --verbose

# JSON output for programmatic processing
./bin/protocol-encoder-go --json --verbose --protocol 0x12345678
```

## License

[MIT License](LICENSE)

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
