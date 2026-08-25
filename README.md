<div align="center">

# Protocol Encoder

[![Arona Version](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Farisu-archive%2Fprotocol-encoder-go%2Frefs%2Fheads%2Fmaster%2Fpkg%2Fencoder%2Farona%2Ftable_gen.go&search=tablegen-version%3A%20%28%5B0-9.%5D%2B%29&replace=v%241&style=for-the-badge&logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgc3Ryb2tlPSIjZmZmIiBzdHJva2Utd2lkdGg9IjIiPjxjaXJjbGUgY3g9IjEyIiBjeT0iMTIiIHI9IjkiLz48cGF0aCBkPSJNMyAxMmgxOE0xMiAzYzQgNSA0IDEzIDAgMThNMTIgM2MtNCA1LTQgMTMgMCAxOCIvPjwvc3ZnPg%3D%3D&label=Arona&color=00ADD8)](pkg/encoder/arona/table_gen.go) [![Plana Version](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Farisu-archive%2Fprotocol-encoder-go%2Frefs%2Fheads%2Fmaster%2Fpkg%2Fencoder%2Fplana%2Ftable_gen.go&search=tablegen-version%3A%20%28%5B0-9.%5D%2B%29&replace=v%241&style=for-the-badge&logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgc3Ryb2tlPSIjZmZmIiBzdHJva2Utd2lkdGg9IjIiPjxwYXRoIGQ9Ik0zIDVoN3YxMGMwIDMtMiA0LTQgNHMtNC0xLTQtNE0xNCAxOVY1aDRjMyAwIDQgMiA0IDRzLTEgNC00IDRoLTQiLz48L3N2Zz4%3D&label=Plana&color=7d3cc8)](pkg/encoder/plana/table_gen.go)
[![Test](https://img.shields.io/github/actions/workflow/status/arisu-archive/protocol-encoder-go/test.yml?branch=master&style=for-the-badge&logo=github&label=Test)](https://github.com/arisu-archive/protocol-encoder-go/actions/workflows/test.yml) [![Latest Release](https://img.shields.io/github/v/release/arisu-archive/protocol-encoder-go?style=for-the-badge&logo=github&label=Release)](https://github.com/arisu-archive/protocol-encoder-go/releases)
[![Go Reference](https://img.shields.io/badge/Go-Reference-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://pkg.go.dev/github.com/arisu-archive/protocol-encoder-go/v2) [![License](https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge)](LICENSE)

**Pure Go protocol encoders for the Nexon and Japanese Yostar versions of Blue Archive.**

[Features](#features) • [Installation](#installation) • [Quick start](#quick-start) • [Encoding behavior](#encoding-behavior) • [Updating generated tables](#updating-generated-tables) • [Documentation](#documentation) • [Testing](#testing) • [Related projects](#related-projects) • [License](#license)

</div>

---

## Features

- Separate typed encoders for the Nexon (`arona`) and Japanese Yostar (`plana`) clients.
- Pure Go runtime with no Unicorn Engine, native library, or `libil2cpp.so` dependency.
- Generated lookup tables derived from each game's latest dispatcher implementation.
- CRC-aware encoding with behavior verified across every generated protocol and route.
- Automatic pass-through for protocols absent from the generated override table.
- Scheduled table updates as new game clients are published.

## Installation

Install the latest release with Go modules:

```bash
go get github.com/arisu-archive/protocol-encoder-go/v2@latest
```

Choose the encoder and protocol package for the matching game service:

| Service | Encoder | Protocol definitions |
| --- | --- | --- |
| Nexon | `github.com/arisu-archive/protocol-encoder-go/v2/pkg/encoder/arona` | `github.com/arisu-archive/arona-protos/protos` |
| Japanese Yostar | `github.com/arisu-archive/protocol-encoder-go/v2/pkg/encoder/plana` | `github.com/arisu-archive/plana-protos/protos` |

## Quick start

Encode protocol identifiers using the package for the target service:

```go
crc := uint32(0x12345678)

nexonValue := arona.Encode(aronaprotos.Protocol_Account_Auth, crc)
japanValue := plana.Encode(planaprotos.Protocol_Account_Auth, crc)
```

The protocol argument is intentionally typed by the corresponding generated protocol package, preventing accidental use of identifiers from the other service.

## Encoding behavior

`Encode` selects a generated value using the protocol identifier and `crc % 99`. CRC values with the same remainder therefore produce the same encoded value.

If a protocol is absent from the generated override table, `Encode` returns its `uint32` representation unchanged. This mirrors the game library's dispatcher behavior and allows newly introduced identity-mapped protocols to pass through safely.

The generated tables are compiled into the packages and are safe to use concurrently. Runtime consumers do not load game binaries or invoke Unicorn Engine.

## Updating generated tables

The [table update workflow](.github/workflows/bump.yml) checks for new Nexon and Japanese Yostar clients hourly on weekdays. For each available update, it downloads the matching `libil2cpp.so` and dispatcher offset, regenerates and verifies the affected table, runs its tests, and opens a pull request.

Table generation lives in the separate `tools` Go module. Unlike the runtime packages, the generator builds the repository's patched Unicorn Engine submodule to execute the ARM64 dispatcher in a controlled environment. Maintainers with the expected inputs under `libraries/` can regenerate either table with:

```bash
make generate-arona
make generate-plana
```

Generated files include the source client version, library digest, dispatcher offset, protocol-set digest, and table-body digest for reproducibility and cache validation. Do not edit them manually.

## Documentation

Browse the package references on pkg.go.dev:

- [`encoder/arona`](https://pkg.go.dev/github.com/arisu-archive/protocol-encoder-go/v2/pkg/encoder/arona) for the Nexon client.
- [`encoder/plana`](https://pkg.go.dev/github.com/arisu-archive/protocol-encoder-go/v2/pkg/encoder/plana) for the Japanese Yostar client.

## Testing

Test and vet the runtime library with:

```bash
make test-race
make vet
```

Generator tests additionally require CMake, Ninja, a C compiler, and the Unicorn submodule:

```bash
make test-tools
```

## Related projects

- [`arona-protos`](https://github.com/arisu-archive/arona-protos) provides Nexon protocol identifiers and models.
- [`plana-protos`](https://github.com/arisu-archive/plana-protos) provides Japanese Yostar protocol identifiers and models.

## License

Protocol Encoder is available under the [MIT License](LICENSE).
