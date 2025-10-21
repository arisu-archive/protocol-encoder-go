package serve

import (
	"fmt"
	"net/http"
	"time"

	"github.com/arisu-archive/protocol-encoder-go/internal/cli"
	"github.com/arisu-archive/protocol-encoder-go/pkg/encoder"
	"github.com/spf13/cobra"
)

const shutdownTimeout = 30 * time.Second

type serveCmd struct {
	port            string
	shutdownTimeout time.Duration
}

func NewCommand() *cobra.Command {
	serveCmd := &serveCmd{}
	cmd := &cobra.Command{
		Use:     "serve",
		Short:   "Serve the protocol encoder",
		Example: "protocol-encoder serve --port :8080",
		RunE:    cli.RunE("serve", serveCmd.execute),
	}

	cmd.Flags().StringVarP(&serveCmd.port, "port", "p", ":8080", "server port")
	cmd.Flags().DurationVarP(&serveCmd.shutdownTimeout, "shutdown-timeout", "s", shutdownTimeout, "shutdown timeout")
	return cmd
}

func (s *serveCmd) execute(cmd *cobra.Command, args []string) error {
	// Get the logger from context.
	logger := cli.GetLogger(cmd.Context())
	if logger == nil {
		return fmt.Errorf("logger is not set in context")
	}

	// Get the config from context.
	cfg := cli.GetConfig(cmd.Context())
	if cfg == nil {
		return fmt.Errorf("config is not set in context")
	}

	// Setup encoder
	encoders := map[string]*encoder.Encoder{}
	for name, encoderCfg := range cfg.Encoders {
		enc, err := cli.SetupEncoder(cfg.Emulator, encoderCfg, logger)
		if err != nil {
			return fmt.Errorf("failed to setup encoder for %s: %w", name, err)
		}
		encoders[name] = enc
		defer enc.Close()
	}

	app := setupWeb(encoders, logger)
	go func() {
		logger.WithField("port", s.port).Info("starting server")
		if err := app.Serve(s.port); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("failed to start server")
		}
	}()
	return app.Shutdown(s.shutdownTimeout)
}
