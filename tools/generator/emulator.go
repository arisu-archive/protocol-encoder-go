package main

import (
	"errors"
	"fmt"
	"math"
	"slices"

	uc "github.com/unicorn-engine/unicorn/bindings/go/unicorn"
)

const (
	imageBase        = uint64(0x10000000)
	pageSize         = uint64(0x1000)
	invocationCount  = routeCount * 2
	stackStride      = uint64(0x10000)
	executionTimeout = uint64(100000)
	instructionLimit = uint64(4096)
)

type extraction struct {
	rows map[uint32]*[routeCount]uint32
}

type emulator struct {
	machine      uc.Unicorn
	context      uc.Context
	entry        uint64
	returnTarget uint64
	stackBase    uint64
	stackSize    uint64
	zeroStack    []byte
}

func extract(image []byte, offset uint64, protocols []uint32) (result extraction, err error) {
	emu, err := newEmulator(image, offset)
	if err != nil {
		return extraction{}, err
	}
	defer func() {
		err = errors.Join(err, emu.close())
	}()

	result.rows = make(map[uint32]*[routeCount]uint32, len(protocols))
	for _, protocol := range protocols {
		if err := emu.resetStack(); err != nil {
			return extraction{}, fmt.Errorf("protocol %d: clear stack: %w", int32(protocol), err)
		}
		var values [routeCount]uint32
		for crc := uint32(0); crc < routeCount; crc++ {
			first, err := emu.invoke(crc, protocol)
			if err != nil {
				return extraction{}, fmt.Errorf("protocol %d crc %d: %w", int32(protocol), crc, err)
			}
			second, err := emu.invoke(crc+routeCount, protocol)
			if err != nil {
				return extraction{}, fmt.Errorf("protocol %d crc %d: %w", int32(protocol), crc+routeCount, err)
			}
			if first != second {
				return extraction{}, fmt.Errorf("protocol %d is not periodic at crc %d: got %d and %d", int32(protocol), crc, first, second)
			}
			values[crc] = first
		}
		if slices.ContainsFunc(values[:], func(value uint32) bool { return value != protocol }) {
			result.rows[protocol] = &values
		}
	}
	return result, nil
}

func newEmulator(image []byte, offset uint64) (_ *emulator, err error) {
	if len(image) == 0 {
		return nil, errors.New("library is empty")
	}
	imageSize := uint64(len(image))
	if imageSize > math.MaxUint64-(pageSize-1) {
		return nil, errors.New("library is too large to map")
	}
	imageSize = (imageSize + pageSize - 1) &^ (pageSize - 1)
	stackSize := uint64(invocationCount) * stackStride
	stackBase := imageBase + imageSize + pageSize
	returnTarget := stackBase + stackSize + pageSize

	machine, err := uc.NewUnicorn(uc.ARCH_ARM64, uc.MODE_ARM)
	if err != nil {
		return nil, fmt.Errorf("create arm64 emulator: %w", err)
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, machine.Close())
		}
	}()
	if err = machine.MemMapProt(imageBase, imageSize, uc.PROT_READ|uc.PROT_WRITE); err != nil {
		return nil, fmt.Errorf("map library: %w", err)
	}
	if err = machine.MemWrite(imageBase, image); err != nil {
		return nil, fmt.Errorf("load library: %w", err)
	}
	if err = machine.MemProtect(imageBase, imageSize, uc.PROT_READ|uc.PROT_EXEC); err != nil {
		return nil, fmt.Errorf("protect library: %w", err)
	}
	if err = machine.MemMapProt(stackBase, stackSize, uc.PROT_READ|uc.PROT_WRITE); err != nil {
		return nil, fmt.Errorf("map stack arena: %w", err)
	}
	registers := []int{uc.ARM64_REG_X0, uc.ARM64_REG_SP, uc.ARM64_REG_LR}
	values := []uint64{0, stackBase + stackStride, returnTarget}
	if err = machine.RegWriteBatch(registers, values); err != nil {
		return nil, fmt.Errorf("initialize registers: %w", err)
	}
	context, err := machine.ContextSave(nil)
	if err != nil {
		return nil, fmt.Errorf("save clean context: %w", err)
	}
	return &emulator{
		machine:      machine,
		context:      context,
		entry:        imageBase + offset,
		returnTarget: returnTarget,
		stackBase:    stackBase,
		stackSize:    stackSize,
		zeroStack:    make([]byte, stackSize),
	}, nil
}

func (e *emulator) resetStack() error {
	return e.machine.MemWrite(e.stackBase, e.zeroStack)
}

func (e *emulator) invoke(crc, protocol uint32) (uint32, error) {
	if err := e.machine.ContextRestore(e.context); err != nil {
		return 0, fmt.Errorf("restore context: %w", err)
	}
	stackPointer := e.stackBase + (uint64(crc)+1)*stackStride
	registers := []int{uc.ARM64_REG_X1, uc.ARM64_REG_X2, uc.ARM64_REG_SP}
	values := []uint64{uint64(crc), uint64(protocol), stackPointer}
	if err := e.machine.RegWriteBatch(registers, values); err != nil {
		return 0, fmt.Errorf("write input registers: %w", err)
	}
	options := uc.UcOptions{Timeout: executionTimeout, Count: instructionLimit}
	if err := e.machine.StartWithOptions(e.entry, e.returnTarget, &options); err != nil {
		return 0, fmt.Errorf("execute dispatcher: %w", err)
	}
	values, err := e.machine.RegReadBatch([]int{uc.ARM64_REG_PC, uc.ARM64_REG_X0})
	if err != nil {
		return 0, fmt.Errorf("read result registers: %w", err)
	}
	if values[0] != e.returnTarget {
		return 0, fmt.Errorf("dispatcher stopped at %#x instead of return target %#x", values[0], e.returnTarget)
	}
	if values[1] > math.MaxUint32 {
		return 0, fmt.Errorf("dispatcher returned non-uint32 value %#x", values[1])
	}
	return uint32(values[1]), nil
}

func (e *emulator) close() error {
	if err := e.machine.Close(); err != nil {
		return fmt.Errorf("close emulator: %w", err)
	}
	return nil
}
