package root

import (
	"io"

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
		logger := cli.GetLogger(r.cmd.Context())
		logger.WithContext(r.cmd.Context()).WithError(err).Error("protocol-encoder failed.")
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

			// Setup config
			cfg, err := cli.InitConfig(root.cfgFile, logger)
			if err != nil {
				logger.WithError(err).Fatal("failed to load config")
			}
			ctx := cli.WithConfig(cli.WithLogger(cmd.Context(), logger), cfg)
			cmd.SetContext(ctx)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			logger := cli.GetLogger(cmd.Context())
			logger.WithContext(cmd.Context()).Info("protocol-encoder completed successfully")
		},
	}
	cmd.PersistentFlags().StringVarP(&root.cfgFile, "config", "c", "", "config file (default is ./config.yaml)")
	cmd.PersistentFlags().BoolVarP(&root.verbose, "verbose", "v", false, "Enable verbose mode")
	cmd.PersistentFlags().BoolVar(&root.jsonFormat, "json", false, "output as JSON")

	// Bind flags to viper
	if err := viper.BindPFlag("cfgFile", cmd.PersistentFlags().Lookup("config")); err != nil {
		panic("Failed to bind config flag to viper")
	}
	if err := viper.BindPFlag("verbose", cmd.PersistentFlags().Lookup("verbose")); err != nil {
		panic("Failed to bind verbose flag to viper")
	}

	cmd.AddCommand(serve.NewCommand())
	cmd.AddCommand(encode.NewCommand())
	cmd.SetIn(in)
	cmd.SetOut(out)
	cmd.SetErr(err)

	root.cmd = cmd
	return root
}
