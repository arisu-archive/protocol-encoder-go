package emulator

import "fmt"

// EmulatorError represents errors that occur during emulation
type EmulatorError struct {
	Operation string
	Err       error
}

func (e *EmulatorError) Error() string {
	return fmt.Sprintf("emulator error during %s: %v", e.Operation, e.Err)
}

func (e *EmulatorError) Unwrap() error {
	return e.Err
}

// NewEmulatorError creates a new emulator error
func NewEmulatorError(operation string, err error) *EmulatorError {
	return &EmulatorError{
		Operation: operation,
		Err:       err,
	}
}

// Memory operation errors
var (
	ErrMemoryMapping    = "memory mapping"
	ErrMemoryWrite      = "memory write"
	ErrRegisterWrite    = "register write"
	ErrRegisterRead     = "register read"
	ErrEmulationStart   = "emulation start"
	ErrEmulationStop    = "emulation stop"
	ErrInitialization   = "initialization"
	ErrFileRead         = "file read"
	ErrOffsetParsing    = "offset parsing"
	ErrConfigValidation = "config validation"
	ErrInvocation       = "function invocation"
)

// Predefined error instances
var (
	ErrNotInitialized  = fmt.Errorf("emulator not initialized")
	ErrBinaryNotLoaded = fmt.Errorf("binary not loaded")
	ErrPoolTimeout     = fmt.Errorf("pool timeout")
)
