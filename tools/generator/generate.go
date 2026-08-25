package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	aronaprotos "github.com/arisu-archive/arona-protos/protos"
	planaprotos "github.com/arisu-archive/plana-protos/protos"
)

const routeCount = 99

func run(cfg config) error {
	offset, empty, err := readOffset(cfg.offsetPath)
	if err != nil {
		return err
	}
	if empty {
		return reportf("skipped: empty offset file %s\n", cfg.offsetPath)
	}

	protocols := protocolIDs(cfg.packageName)
	protocolHash := hashProtocols(protocols)
	image, err := os.ReadFile(cfg.libraryPath)
	if err != nil {
		return fmt.Errorf("read library %q: %w", cfg.libraryPath, err)
	}
	if offset >= uint64(len(image)) || offset%4 != 0 {
		return fmt.Errorf("dispatcher offset %#x is outside the library or unaligned", offset)
	}

	meta := metadata{
		packageName:  cfg.packageName,
		version:      cfg.version,
		libraryHash:  sha256.Sum256(image),
		offset:       offset,
		protocolHash: protocolHash,
	}
	prefix := meta.headerPrefix()
	current, err := readOutput(cfg.outputPath, cfg.check)
	if err != nil {
		return err
	}
	if validCache(current, prefix) {
		if cfg.check {
			return reportf("verified %s\n", cfg.outputPath)
		}
		return reportf("cache hit: %s\n", cfg.outputPath)
	}

	result, err := extract(image, offset, protocols)
	if err != nil {
		return fmt.Errorf("extract %s table: %w", cfg.packageName, err)
	}
	source, err := render(cfg, meta, protocols, result)
	if err != nil {
		return err
	}
	if cfg.check {
		if !bytes.Equal(current, source) {
			return fmt.Errorf("generated output is stale: %s", cfg.outputPath)
		}
		return reportf("verified %s\n", cfg.outputPath)
	}
	if err := writeAtomic(cfg.outputPath, source); err != nil {
		return err
	}
	return reportf("wrote %s\n", cfg.outputPath)
}

func readOffset(path string) (uint64, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false, fmt.Errorf("read dispatcher offset %q: %w", path, err)
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		return 0, true, nil
	}
	offset, err := strconv.ParseUint(text, 0, 64)
	if err != nil {
		return 0, false, fmt.Errorf("parse dispatcher offset %q: %w", text, err)
	}
	return offset, false, nil
}

func protocolIDs(packageName string) []uint32 {
	var protocols []uint32
	if packageName == "arona" {
		values := aronaprotos.ProtocolValues()
		protocols = make([]uint32, len(values))
		for i, protocol := range values {
			protocols[i] = uint32(protocol)
		}
	} else {
		values := planaprotos.ProtocolValues()
		protocols = make([]uint32, len(values))
		for i, protocol := range values {
			protocols[i] = uint32(protocol)
		}
	}
	slices.Sort(protocols)
	return slices.Compact(protocols)
}

func hashProtocols(protocols []uint32) [sha256.Size]byte {
	data := make([]byte, 4+len(protocols)*4)
	binary.LittleEndian.PutUint32(data, uint32(len(protocols)))
	for i, protocol := range protocols {
		binary.LittleEndian.PutUint32(data[4+i*4:], protocol)
	}
	return sha256.Sum256(data)
}

func reportf(formatText string, args ...any) error {
	if _, err := fmt.Fprintf(os.Stdout, formatText, args...); err != nil {
		return fmt.Errorf("write status: %w", err)
	}
	return nil
}
