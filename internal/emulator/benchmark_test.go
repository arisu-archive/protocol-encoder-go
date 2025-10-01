package emulator

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// BenchmarkSingleInvoke benchmarks a single function invocation
func BenchmarkSingleInvoke(b *testing.B) {
	// Skip if test files don't exist
	if _, err := os.Stat("../../offset.txt"); os.IsNotExist(err) {
		b.Skip("Skipping benchmark - offset.txt not found")
	}
	if _, err := os.Stat("../../libil2cpp.so"); os.IsNotExist(err) {
		b.Skip("Skipping benchmark - libil2cpp.so not found")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during benchmark

	config := DefaultConfig()
	emu, err := New(config, logger)
	require.NoError(b, err)
	defer emu.Close()

	require.NoError(b, emu.Initialize())
	require.NoError(b, emu.Load("../../libil2cpp.so"))

	offset := uint64(0x6268754) // Example offset

	req := &InvokeRequest{
		Offset:  offset,
		Args:    []uint64{0, 1, 1014}, // X0=0, X1=1, X2=1014
		Context: context.Background(),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := emu.Invoke(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPoolInvoke benchmarks function invocation using emulator pool
func BenchmarkPoolInvoke(b *testing.B) {
	// Skip if test files don't exist
	if _, err := os.Stat("../../offset.txt"); os.IsNotExist(err) {
		b.Skip("Skipping benchmark - offset.txt not found")
	}
	if _, err := os.Stat("../../libil2cpp.so"); os.IsNotExist(err) {
		b.Skip("Skipping benchmark - libil2cpp.so not found")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during benchmark

	config := DefaultConfig()
	pool, err := NewPool(config, logger, 4) // 4 emulators in pool
	require.NoError(b, err)
	defer pool.Close()

	// Load binary into all emulators in the pool
	for i := 0; i < 4; i++ {
		emu := <-pool.pool
		require.NoError(b, emu.Load("../../libil2cpp.so"))
		pool.pool <- emu
	}

	offset := uint64(0x6268754) // Example offset

	req := &InvokeRequest{
		Offset:  offset,
		Args:    []uint64{0, 1, 1014}, // X0=0, X1=1, X2=1014
		Context: context.Background(),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := pool.Invoke(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParallelPoolInvoke benchmarks parallel function invocations
func BenchmarkParallelPoolInvoke(b *testing.B) {
	// Skip if test files don't exist
	if _, err := os.Stat("../../offset.txt"); os.IsNotExist(err) {
		b.Skip("Skipping benchmark - offset.txt not found")
	}
	if _, err := os.Stat("../../libil2cpp.so"); os.IsNotExist(err) {
		b.Skip("Skipping benchmark - libil2cpp.so not found")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := DefaultConfig()
	pool, err := NewPool(config, logger, 8) // 8 emulators for parallel test
	require.NoError(b, err)
	defer pool.Close()

	// Load binary into all emulators
	for i := 0; i < 8; i++ {
		emu := <-pool.pool
		require.NoError(b, emu.Load("../../libil2cpp.so"))
		pool.pool <- emu
	}

	offset := uint64(0x6268754) // Example offset

	req := &InvokeRequest{
		Offset:  offset,
		Args:    []uint64{0, 1, 1014}, // X0=0, X1=1, X2=1014
		Context: context.Background(),
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := pool.Invoke(req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := DefaultConfig()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		emu, err := New(config, logger)
		if err != nil {
			b.Fatal(err)
		}
		emu.Close()
	}
}

// BenchmarkContextCancellation benchmarks context cancellation handling
func BenchmarkContextCancellation(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := DefaultConfig()
	emu, err := New(config, logger)
	require.NoError(b, err)
	defer emu.Close()

	require.NoError(b, emu.Initialize())

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		req := &InvokeRequest{
			Offset:  0x1000,
			Args:    []uint64{0, 1, 1014}, // X0=0, X1=1, X2=1014
			Context: ctx,
		}

		// This should be cancelled immediately
		_, _ = emu.Invoke(req)
		cancel()
	}
}
