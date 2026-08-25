package plana

import protos "github.com/arisu-archive/plana-protos/protos"

const routeCount = 99

// Encode returns the encoded value for protocol and crc. Protocols absent from
// the generated override table are returned unchanged after uint32 conversion.
func Encode(protocol protos.Protocol, crc uint32) uint32 {
	row, ok := protocolTable[protocol]
	if !ok {
		return uint32(protocol)
	}
	return row[crc%routeCount]
}
