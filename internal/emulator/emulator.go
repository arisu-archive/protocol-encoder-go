package emulator

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicorn-engine/unicorn/bindings/go/unicorn"
)

// Emulator represents an ARM emulator instance
type Emulator struct {
	config       *Config
	engine       unicorn.Unicorn
	logger       *logrus.Logger
	binary       []byte
	binaryLoaded bool
	initialized  bool
	mu           sync.RWMutex
}

// InvokeRequest represents a function invocation request
type InvokeRequest struct {
	Offset  uint64          `json:"offset"`
	Args    []uint64        `json:"args"` // All function arguments (including 'this' pointer if applicable)
	Context context.Context `json:"-"`
}

// InvokeResponse represents the result of function invocation
type InvokeResponse struct {
	ReturnValue     uint64            `json:"return_value"`
	FunctionAddress uint64            `json:"function_address"`
	ExecutionTime   time.Duration     `json:"execution_time"`
	Success         bool              `json:"success"`
	Error           string            `json:"error,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// New creates a new emulator instance
func New(config *Config, logger *logrus.Logger) (*Emulator, error) {
	if err := config.Validate(); err != nil {
		return nil, NewEmulatorError(ErrConfigValidation, err)
	}

	return &Emulator{
		config: config,
		logger: logger,
	}, nil
}

// Initialize sets up the emulator engine
func (e *Emulator) Initialize() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized {
		return nil // Already initialized
	}

	e.logger.Debug("Initializing ARM emulator")

	// Initialize Unicorn engine
	if err := e.initializeEngine(); err != nil {
		return err
	}

	e.initialized = true
	e.logger.Debug("Emulator initialized successfully")
	return nil
}

// Load loads a binary file into the emulator
func (e *Emulator) Load(binaryPath string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.initialized {
		return NewEmulatorError(ErrInitialization, ErrNotInitialized)
	}

	e.logger.WithField("binary", binaryPath).Debug("Loading binary")

	// Load binary file
	binary, err := os.ReadFile(binaryPath)
	if err != nil {
		return NewEmulatorError(ErrFileRead, err)
	}

	e.binary = binary
	e.binaryLoaded = false // Reset loaded state

	// Setup memory with new binary
	if err := e.setupMemory(); err != nil {
		return err
	}

	e.binaryLoaded = true
	e.logger.WithField("binary_size", len(e.binary)).Info("Binary loaded successfully")
	return nil
}

// Invoke executes a function at the specified offset with given arguments
func (e *Emulator) Invoke(req *InvokeRequest) (*InvokeResponse, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.binaryLoaded {
		return nil, NewEmulatorError(ErrInvocation, ErrBinaryNotLoaded)
	}

	startTime := time.Now()

	// Check context cancellation
	if req.Context != nil {
		select {
		case <-req.Context.Done():
			return &InvokeResponse{
				Success: false,
				Error:   "context cancelled",
			}, req.Context.Err()
		default:
		}
	}

	// Setup registers for this invocation
	if err := e.setupRegistersForInvoke(req); err != nil {
		return &InvokeResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Calculate function address
	funcAddr := e.config.BaseAddr + req.Offset

	// Start emulation
	err := e.engine.Start(funcAddr, e.config.ReturnAddr)
	executionTime := time.Since(startTime)

	if err != nil {
		e.logger.WithError(err).Error("Emulation failed")
		return &InvokeResponse{
			FunctionAddress: funcAddr,
			ExecutionTime:   executionTime,
			Success:         false,
			Error:           err.Error(),
		}, NewEmulatorError(ErrEmulationStart, err)
	}

	// Read result
	result, err := e.engine.RegRead(unicorn.ARM64_REG_X0)
	if err != nil {
		return &InvokeResponse{
			FunctionAddress: funcAddr,
			ExecutionTime:   executionTime,
			Success:         false,
			Error:           err.Error(),
		}, NewEmulatorError(ErrRegisterRead, err)
	}

	response := &InvokeResponse{
		ReturnValue:     result,
		FunctionAddress: funcAddr,
		ExecutionTime:   executionTime,
		Success:         true,
	}

	e.logger.WithFields(logrus.Fields{
		"return_value":     result,
		"function_address": funcAddr,
		"execution_time":   executionTime,
	}).Debug("Function invocation completed successfully")

	return response, nil
}

// Close cleans up the emulator resources
func (e *Emulator) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.engine != nil {
		return e.engine.Close()
	}
	return nil
}

// initializeEngine sets up the Unicorn engine
func (e *Emulator) initializeEngine() error {
	e.logger.Debug("Initializing Unicorn engine")

	engine, err := unicorn.NewUnicorn(unicorn.ARCH_ARM64, unicorn.MODE_ARM)
	if err != nil {
		return NewEmulatorError(ErrInitialization, err)
	}

	e.engine = engine
	return nil
}

// setupMemory maps memory regions and loads the binary
func (e *Emulator) setupMemory() error {
	e.logger.Debug("Setting up memory regions")

	// Load the ENTIRE binary exactly like the Python POC
	binarySize := len(e.binary)

	// Calculate mapped size (round up to page size 4KB)
	mappedSize := uint64(((binarySize + 0xFFF) / 0x1000) * 0x1000)

	e.logger.WithFields(logrus.Fields{
		"binary_size": binarySize,
		"mapped_size": mappedSize,
	}).Debug("Mapping ENTIRE binary")

	// Map binary memory for the FULL size
	err := e.engine.MemMap(e.config.BaseAddr, mappedSize)
	if err != nil {
		return NewEmulatorError(ErrMemoryMapping, err)
	}

	// Map stack memory
	err = e.engine.MemMap(e.config.StackAddr, e.config.StackSize)
	if err != nil {
		return NewEmulatorError(ErrMemoryMapping, err)
	}

	// Write the ENTIRE binary to memory
	err = e.engine.MemWrite(e.config.BaseAddr, e.binary)
	if err != nil {
		return NewEmulatorError(ErrMemoryWrite, err)
	}

	e.logger.WithFields(logrus.Fields{
		"base_addr":  e.config.BaseAddr,
		"stack_addr": e.config.StackAddr,
		"stack_size": e.config.StackSize,
	}).Debug("Memory setup completed")

	return nil
}

// setupRegistersForInvoke configures the ARM64 registers for a specific function invocation
func (e *Emulator) setupRegistersForInvoke(req *InvokeRequest) error {
	// Set stack pointer
	stackPointer := e.config.StackAddr + e.config.StackSize - 0x1000
	if err := e.engine.RegWrite(unicorn.ARM64_REG_SP, stackPointer); err != nil {
		return NewEmulatorError(ErrRegisterWrite, err)
	}

	// Clear all argument registers first
	armRegs := []int{
		unicorn.ARM64_REG_X0,
		unicorn.ARM64_REG_X1,
		unicorn.ARM64_REG_X2,
		unicorn.ARM64_REG_X3,
		unicorn.ARM64_REG_X4,
		unicorn.ARM64_REG_X5,
		unicorn.ARM64_REG_X6,
		unicorn.ARM64_REG_X7,
	}

	// Clear all registers first
	for _, reg := range armRegs {
		if err := e.engine.RegWrite(reg, 0); err != nil {
			return NewEmulatorError(ErrRegisterWrite, err)
		}
	}

	// Set function arguments starting from X0
	// This supports any calling convention:
	// - thiscall: X0 = this, X1+ = args
	// - fastcall: X0+ = args
	// - stdcall: X0+ = args
	// - custom: user defines what goes where
	for i, arg := range req.Args {
		if i < len(armRegs) {
			if err := e.engine.RegWrite(armRegs[i], arg); err != nil {
				return NewEmulatorError(ErrRegisterWrite, err)
			}
		} else {
			// If more than 8 arguments, they would go on stack
			// For ARM64 AAPCS, arguments beyond X7 go on stack
			stackOffset := uint64((i - len(armRegs)) * 8)   // 8 bytes per argument
			stackAddr := stackPointer - 0x100 + stackOffset // Leave some space at top of stack

			// Write argument to stack
			argBytes := make([]byte, 8)
			for j := 0; j < 8; j++ {
				argBytes[j] = byte(arg >> (j * 8))
			}

			if err := e.engine.MemWrite(stackAddr, argBytes); err != nil {
				e.logger.WithFields(logrus.Fields{
					"arg_index":  i,
					"stack_addr": stackAddr,
					"arg_value":  arg,
				}).Warn("Failed to write argument to stack, continuing with register-only arguments")
			} else {
				e.logger.WithFields(logrus.Fields{
					"arg_index":  i,
					"stack_addr": stackAddr,
					"arg_value":  arg,
				}).Debug("Argument written to stack")
			}
		}
	}

	// Set return address
	if err := e.engine.RegWrite(unicorn.ARM64_REG_X30, e.config.ReturnAddr); err != nil {
		return NewEmulatorError(ErrRegisterWrite, err)
	}

	e.logger.WithFields(logrus.Fields{
		"num_args":    len(req.Args),
		"reg_args":    len(armRegs),
		"stack_args":  max(0, len(req.Args)-len(armRegs)),
		"return_addr": e.config.ReturnAddr,
		"stack_ptr":   stackPointer,
	}).Debug("Registers configured for function invocation")

	return nil
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
