package root

import (
	"io"
	"log/slog"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/arisu-archive/protocol-encoder-go/internal/cli"
	"github.com/arisu-archive/protocol-encoder-go/internal/cli/encode"
	"github.com/arisu-archive/protocol-encoder-go/internal/cli/serve"
)

type rootCmd struct {
	cmd        *cobra.Command
	exit       func(int)
	verbose    bool
	cfgFile    string
	jsonFormat bool
}

func Execute(version string, exit func(int), in io.Reader, out, err io.Writer, args []string) {
	newRootCmd(version, exit, in, out, err).Execute(args)
}

func (r *rootCmd) Execute(args []string) {
	r.cmd.SetArgs(args)
	if err := r.cmd.Execute(); err != nil {
		slog.Error("protocol-encoder failed.", slog.Any("error", err))
		r.exit(1)
	}
}

func newRootCmd(version string, exit func(int), in io.Reader, out, err io.Writer) *rootCmd {
	root := &rootCmd{
		exit: exit,
	}

	cmd := &cobra.Command{
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Use:               "protocol-encoder <command> [flags]",
		Short:             "Protocol encoder using Unicorn Engine",
		Long:              `A high-performance protocol encoder that uses Unicorn Engine to encode protocol data using ARM64 binary functions in a controlled emulation environment.`,
		Example:           "protocol-encoder encode --protocol 0xDEADBEEF --crc32 0x12345678",
		Version:           version,
		SilenceErrors:     false,
		SilenceUsage:      false,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger := logrus.New()
			logger.SetLevel(logrus.InfoLevel)
			if root.verbose {
				logger.SetLevel(logrus.DebugLevel)
				logger.Debug("verbose mode enabled")
			}
			if root.jsonFormat {
				logger.SetFormatter(&logrus.JSONFormatter{})
				logger.Debug("json log format enabled")
			}
			cmd.SetContext(cli.WithLogger(cmd.Context(), logger))

			// Setup config
			cli.InitConfig(root.cfgFile, logger)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			logger := cli.GetLogger(cmd.Context())
			logger.WithContext(cmd.Context()).Info("protocol-encoder completed successfully")
		},
	}
	cmd.PersistentFlags().String("binary", "libil2cpp.so", "path to binary file")
	cmd.PersistentFlags().Int("pool", 1, "emulator pool size for high throughput (1 = single emulator, >1 = pool mode)")
	cmd.PersistentFlags().StringVarP(&root.cfgFile, "config", "c", "", "config file (default is ./config.yaml)")
	cmd.PersistentFlags().BoolVarP(&root.verbose, "verbose", "v", false, "Enable verbose mode")
	cmd.PersistentFlags().StringP("offset", "o", "0xDEADBEEF", "function offset (hex or decimal)")
	cmd.PersistentFlags().BoolVar(&root.jsonFormat, "json", false, "output as JSON")

	// Bind flags to viper
	if err := viper.BindPFlag("offset", cmd.PersistentFlags().Lookup("offset")); err != nil {
		panic("Failed to bind offset flag to viper")
	}
	if err := viper.BindPFlag("cfgFile", cmd.PersistentFlags().Lookup("config")); err != nil {
		panic("Failed to bind config flag to viper")
	}
	if err := viper.BindPFlag("verbose", cmd.PersistentFlags().Lookup("verbose")); err != nil {
		panic("Failed to bind verbose flag to viper")
	}
	if err := viper.BindPFlag("pool", cmd.PersistentFlags().Lookup("pool")); err != nil {
		panic("Failed to bind pool flag to viper")
	}
	if err := viper.BindPFlag("binary", cmd.PersistentFlags().Lookup("binary")); err != nil {
		panic("Failed to bind binary flag to viper")
	}

	cmd.AddCommand(serve.NewCommand())
	cmd.AddCommand(encode.NewCommand())
	cmd.SetIn(in)
	cmd.SetOut(out)
	cmd.SetErr(err)

	root.cmd = cmd
	return root
}
