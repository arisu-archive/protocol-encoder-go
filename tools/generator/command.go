package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type config struct {
	packageName string
	libraryPath string
	offsetPath  string
	version     string
	outputPath  string
	check       bool
}

func newCommand() *cobra.Command {
	var cfg config
	cmd := &cobra.Command{
		Use:           "protocol-tablegen",
		Short:         "Generate a protocol lookup table",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			resolved, err := cfg.resolve()
			if err != nil {
				return err
			}
			return run(resolved)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&cfg.packageName, "package", "", "output package: arona or plana")
	flags.StringVar(&cfg.libraryPath, "library", "", "path to libil2cpp.so")
	flags.StringVar(&cfg.offsetPath, "offset", "", "path to dispatcher offset file")
	flags.StringVar(&cfg.version, "version", "", "game version")
	flags.StringVar(&cfg.outputPath, "output", "", "generated Go file")
	flags.BoolVar(&cfg.check, "check", false, "verify output without writing")
	return cmd
}

func (cfg config) resolve() (config, error) {
	if cfg.packageName == "" && cfg.outputPath == "" {
		return config{}, errors.New("either --package or --output is required")
	}
	if cfg.outputPath != "" {
		outputPackage := filepath.Base(filepath.Dir(filepath.Clean(cfg.outputPath)))
		if cfg.packageName == "" {
			cfg.packageName = outputPackage
		} else if cfg.packageName != outputPackage {
			return config{}, fmt.Errorf("package %q does not match output directory %q", cfg.packageName, outputPackage)
		}
	}
	if cfg.packageName != "arona" && cfg.packageName != "plana" {
		return config{}, fmt.Errorf("unsupported package %q: want arona or plana", cfg.packageName)
	}
	if cfg.outputPath == "" {
		root, err := repositoryRoot()
		if err != nil {
			return config{}, err
		}
		cfg.outputPath = filepath.Join(root, cfg.packageName, "table_gen.go")
	}
	if cfg.libraryPath == "" || cfg.offsetPath == "" || strings.TrimSpace(cfg.version) == "" {
		return config{}, errors.New("--library, --offset, and --version are required")
	}
	cfg.version = strings.TrimSpace(cfg.version)
	if strings.ContainsAny(cfg.version, "\r\n") {
		return config{}, errors.New("--version must not contain a newline")
	}
	return cfg, nil
}

func repositoryRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	for {
		if regularFile(filepath.Join(dir, "go.mod")) && regularFile(filepath.Join(dir, "tools", "go.mod")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("cannot find repository root containing go.mod and tools/go.mod")
		}
		dir = parent
	}
}

func regularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
