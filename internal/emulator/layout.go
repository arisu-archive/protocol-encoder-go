package emulator

import "fmt"

const (
	// pageSize is the Unicorn mapping granularity; MemMap rejects unaligned
	// addresses and sizes with UC_ERR_ARG.
	pageSize = 0x1000

	// stackAlignment keeps a relocated stack on a 64KB boundary so its address
	// stays readable in logs and traces.
	stackAlignment = 0x10000

	// guardSize is left unmapped between the image and a relocated stack so a
	// stack overflow faults instead of silently overwriting mapped code.
	guardSize = pageSize
)

// memoryLayout describes the guest address space of a loaded binary.
type memoryLayout struct {
	// ImageSize is the page-aligned size mapped at Config.BaseAddr.
	ImageSize uint64
	// StackAddr is the address the stack is mapped at. It equals the
	// configured stack address unless that region overlaps the image.
	StackAddr uint64
	// Relocated reports whether StackAddr differs from the configured address.
	Relocated bool
}

// resolveMemoryLayout places the image and the stack in non-overlapping guest
// regions. preferredStackAddr is kept whenever it does not collide with the
// image; otherwise the stack moves above the image, past a guard page. It
// returns an error when the requested regions cannot fit in the address space.
func resolveMemoryLayout(baseAddr, binarySize, preferredStackAddr, stackSize uint64) (memoryLayout, error) {
	if binarySize == 0 {
		return memoryLayout{}, fmt.Errorf("binary is empty")
	}
	if stackSize == 0 {
		return memoryLayout{}, fmt.Errorf("stack size must be non-zero")
	}

	imageSize, err := alignUp(binarySize, pageSize)
	if err != nil {
		return memoryLayout{}, fmt.Errorf("binary size %d cannot be page aligned: %w", binarySize, err)
	}

	imageEnd, err := add(baseAddr, imageSize)
	if err != nil {
		return memoryLayout{}, fmt.Errorf("image of %d bytes does not fit at base address 0x%X: %w", imageSize, baseAddr, err)
	}

	layout := memoryLayout{ImageSize: imageSize, StackAddr: preferredStackAddr}

	stackEnd, err := add(preferredStackAddr, stackSize)
	if err != nil {
		return memoryLayout{}, fmt.Errorf("stack of %d bytes does not fit at address 0x%X: %w", stackSize, preferredStackAddr, err)
	}

	if overlaps(baseAddr, imageEnd, preferredStackAddr, stackEnd) {
		relocated, err := add(imageEnd, guardSize)
		if err != nil {
			return memoryLayout{}, fmt.Errorf("no room for a guard page above image end 0x%X: %w", imageEnd, err)
		}
		relocated, err = alignUp(relocated, stackAlignment)
		if err != nil {
			return memoryLayout{}, fmt.Errorf("relocated stack address 0x%X cannot be aligned: %w", relocated, err)
		}
		if _, err := add(relocated, stackSize); err != nil {
			return memoryLayout{}, fmt.Errorf("stack of %d bytes does not fit above image end 0x%X: %w", stackSize, imageEnd, err)
		}

		layout.StackAddr = relocated
		layout.Relocated = true
	}

	return layout, nil
}

// stackPointer returns the initial stack pointer for the layout, one page below
// the top of the stack region.
func (l memoryLayout) stackPointer(stackSize uint64) uint64 {
	return l.StackAddr + stackSize - pageSize
}

func overlaps(startA, endA, startB, endB uint64) bool {
	return startA < endB && startB < endA
}

func alignUp(value, alignment uint64) (uint64, error) {
	remainder := value % alignment
	if remainder == 0 {
		return value, nil
	}
	return add(value, alignment-remainder)
}

func add(a, b uint64) (uint64, error) {
	sum := a + b
	if sum < a {
		return 0, fmt.Errorf("address space overflow adding 0x%X to 0x%X", b, a)
	}
	return sum, nil
}
