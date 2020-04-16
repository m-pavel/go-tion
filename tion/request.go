package tion

import (
	"bytes"
)

var (
	// StatusRequest bytes
	StatusRequest = []byte{0x3d, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5a}
)

// FromStatus to bytes
func FromStatus(s *Status) []byte {
	return BuildRequest(s.Enabled, s.SoundEnabled, s.HeaterEnabled, s.Speed, s.Gate, s.TempTarget, s.ResetFilters)
}

// BuildRequest with given parameters
func BuildRequest(enabled, sound, heater bool, speed, gate, temp int8, filterReset bool) []byte {
	bf := bytes.NewBufferString("")
	bf.WriteByte(0x3d)
	bf.WriteByte(0x02)
	bf.WriteByte(byte(speed))
	bf.WriteByte(byte(temp))
	bf.WriteByte(byte(gate))
	flags := byte(0)
	if heater {
		flags |= 1
	}
	if enabled {
		flags |= 2
	}
	if sound {
		flags |= 8
	}

	bf.WriteByte(flags)

	if heater {
		bf.WriteByte(0x01)
	} else {
		bf.WriteByte(0x00)
	}

	if filterReset {
		bf.WriteByte(0x02)
	} else {
		bf.WriteByte(0x00)
	}
	bf.Write([]byte{0x0, 0x0, 0x0})
	//	bf.Write([]byte{0x00, 0x00, 0x00, 0x00})             // 4
	bf.Write([]byte{0x00, 0x00})                         // 2
	bf.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) // 6
	bf.WriteByte(0x5a)
	return bf.Bytes()
}
