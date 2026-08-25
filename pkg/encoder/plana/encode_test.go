package plana

import (
	"math"
	"testing"

	protos "github.com/arisu-archive/plana-protos/protos"
)

func TestEncodeUsesProtocolTable(t *testing.T) {
	for protocol, row := range protocolTable {
		if row == nil {
			t.Fatalf("protocolTable[%d] = nil, want override row", protocol)
		}
		for crc := uint32(0); crc < routeCount; crc++ {
			want := row[crc]
			if got := Encode(protocol, crc); got != want {
				t.Fatalf("Encode(%d, %d) = %d, want %d", protocol, crc, got, want)
			}
		}
	}
}

func TestEncodeReturnsUnknownProtocol(t *testing.T) {
	for _, protocol := range []protos.Protocol{-1, math.MaxInt32} {
		if got, want := Encode(protocol, math.MaxUint32), uint32(protocol); got != want {
			t.Errorf("Encode(%d, %d) = %d, want %d", protocol, uint32(math.MaxUint32), got, want)
		}
	}
}

func TestEncodeEquivalentCRCs(t *testing.T) {
	for protocol := range protocolTable {
		for crc := uint32(0); crc < routeCount; crc++ {
			if got, want := Encode(protocol, crc+routeCount), Encode(protocol, crc); got != want {
				t.Fatalf("Encode(%d, %d) = %d, want %d", protocol, crc+routeCount, got, want)
			}
		}
	}
}
