package cli

import (
	"fmt"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

type CommandExecutor func(cmd *cobra.Command, args []string) (err error)

type Command interface {
	Command() *cobra.Command
}

func RunE(verb string, fn func(cmd *cobra.Command, args []string) error) CommandExecutor {
	return func(cmd *cobra.Command, args []string) (err error) {
		start := time.Now()
		defer func() {
			if r := recover(); r != nil {
				// Printout the stack trace
				buf := make([]byte, 1024*1024)
				runtime.Stack(buf, true)
				fmt.Printf("panic: %v\n%s", r, buf)
				err = fmt.Errorf("%s failed after %s: %v", verb, time.Since(start).Truncate(time.Second), r)
			}
		}()
		if err = fn(cmd, args); err != nil {
			return fmt.Errorf("%s failed after %s: %w", verb, time.Since(start).Truncate(time.Second), err)
		}
		logger := GetLogger(cmd.Context())
		logger.WithField("duration", time.Since(start).Truncate(time.Millisecond)).Info(fmt.Sprintf("%s completed", verb))
		return nil
	}
}
