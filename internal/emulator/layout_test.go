package emulator

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resolveMemoryLayout is exercised directly because reaching it through Load
// requires a binary large enough to collide with the stack, which only the
// integration test in emulator_test.go provides.
func TestResolveMemoryLayout(t *testing.T) {
	const (
		baseAddr  = 0x10000000
		stackAddr = 0x20000000
		stackSize = 1024 * 1024
	)

	t.Run("keeps the configured stack when the image does not reach it", func(t *testing.T) {
		layout, err := resolveMemoryLayout(baseAddr, 114569200, stackAddr, stackSize)
		require.NoError(t, err)

		assert.Equal(t, uint64(0x6D43000), layout.ImageSize)
		assert.Equal(t, uint64(stackAddr), layout.StackAddr)
		assert.False(t, layout.Relocated)
	})

	t.Run("relocates the stack when the image covers the configured address", func(t *testing.T) {
		// Size of com.YostarJP.BlueArchive libil2cpp.so 1.71.449178, which maps
		// to 0x20AA7000 and swallows the configured stack address.
		const binarySize = 279601880

		layout, err := resolveMemoryLayout(baseAddr, binarySize, stackAddr, stackSize)
		require.NoError(t, err)

		imageEnd := uint64(baseAddr) + layout.ImageSize
		assert.True(t, layout.Relocated)
		assert.Equal(t, uint64(0x20AA7000), imageEnd)
		assert.Greater(t, layout.StackAddr, imageEnd, "stack must start above the image")
		assert.GreaterOrEqual(t, layout.StackAddr-imageEnd, uint64(guardSize), "guard page must stay unmapped")
		assert.Zero(t, layout.StackAddr%pageSize, "stack address must be page aligned")
		assert.Greater(t, layout.stackPointer(stackSize), layout.StackAddr)
	})

	t.Run("relocates the stack when the image ends inside the stack region", func(t *testing.T) {
		// Image end lands one page into the stack region.
		binarySize := uint64(stackAddr - baseAddr + pageSize)

		layout, err := resolveMemoryLayout(baseAddr, binarySize, stackAddr, stackSize)
		require.NoError(t, err)

		imageEnd := uint64(baseAddr) + layout.ImageSize
		assert.True(t, layout.Relocated)
		assert.GreaterOrEqual(t, layout.StackAddr-imageEnd, uint64(guardSize))
	})

	t.Run("keeps the configured stack when the image ends exactly at it", func(t *testing.T) {
		layout, err := resolveMemoryLayout(baseAddr, stackAddr-baseAddr, stackAddr, stackSize)
		require.NoError(t, err)

		assert.Equal(t, uint64(stackAddr), layout.StackAddr)
		assert.False(t, layout.Relocated)
	})

	t.Run("rejects an empty binary", func(t *testing.T) {
		_, err := resolveMemoryLayout(baseAddr, 0, stackAddr, stackSize)
		assert.ErrorContains(t, err, "binary is empty")
	})

	t.Run("rejects a zero stack size", func(t *testing.T) {
		_, err := resolveMemoryLayout(baseAddr, pageSize, stackAddr, 0)
		assert.ErrorContains(t, err, "stack size must be non-zero")
	})

	t.Run("rejects an image that overflows the address space", func(t *testing.T) {
		_, err := resolveMemoryLayout(math.MaxUint64-pageSize+1, pageSize, stackAddr, stackSize)
		assert.ErrorContains(t, err, "does not fit at base address")
	})

	t.Run("rejects a stack that overflows the address space", func(t *testing.T) {
		_, err := resolveMemoryLayout(baseAddr, pageSize, math.MaxUint64-pageSize+1, stackSize)
		assert.ErrorContains(t, err, "does not fit at address")
	})

	t.Run("rejects a relocated stack that no longer fits", func(t *testing.T) {
		// The configured stack overlaps the image and fits on its own, but the
		// image ends too close to the top of the address space to relocate it.
		const (
			nearTop          = 0xFFFFFFFFFFE00000
			overlappingStack = 0xFFFFFFFFFFE80000
			largeStackSize   = 0x100000
			imageReachingTop = 0x100000
		)

		_, err := resolveMemoryLayout(nearTop, imageReachingTop, overlappingStack, largeStackSize)
		assert.ErrorContains(t, err, "does not fit above image end")
	})
}
