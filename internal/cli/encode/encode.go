package encode

import (
	"fmt"

	"github.com/arisu-archive/protocol-encoder-go/internal/cli"
	"github.com/spf13/cobra"
)

type encodeCmd struct {
	server   string
	crc32    uint64
	protocol uint64
}

func NewCommand() *cobra.Command {
	encodeCmd := &encodeCmd{}
	cmd := &cobra.Command{
		Use:     "encode",
		Short:   "Encode a protocol",
		Example: "protocol-encoder encode --crc32 0x12345678 --protocol 0xDEADBEEF",
		RunE:    cli.RunE("encode", encodeCmd.execute),
	}
	cmd.Flags().StringVarP(&encodeCmd.server, "server", "s", "global", "server address to send the request to (if not set, runs locally)")
	cmd.Flags().Uint64Var(&encodeCmd.crc32, "crc32", 0, "packet crc32 value")
	cmd.Flags().Uint64Var(&encodeCmd.protocol, "protocol", 0xDEADBEEF, "protocol value to encode")
	return cmd
}

func (e *encodeCmd) execute(cmd *cobra.Command, args []string) error {
	// Get the logger from context.
	logger := cli.GetLogger(cmd.Context())
	if logger == nil {
		return fmt.Errorf("logger is not set in context")
	}
	cfg := cli.GetConfig(cmd.Context())
	if cfg == nil {
		return fmt.Errorf("config is not set in context")
	}

	// If server is set, run in client mode.
	encoderCfg, ok := cfg.Encoders[e.server]
	if !ok {
		return fmt.Errorf("server %s not found in config", e.server)
	}
	// Setup encoder
	enc, err := cli.SetupEncoder(cfg.Emulator, encoderCfg, logger)
	if err != nil {
		return fmt.Errorf("failed to setup encoder: %w", err)
	}
	defer enc.Close()

	res, err := enc.Encode(e.protocol, e.crc32)
	if err != nil {
		return fmt.Errorf("failed to encode: %w", err)
	}
	logger.WithField("encoded", res.ReturnValue).Info("Encoding successful")

	return nil
}
