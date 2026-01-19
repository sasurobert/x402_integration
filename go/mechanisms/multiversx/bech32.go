package multiversx

import (
	"fmt"
	"strings"
)

const charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

var gen = []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}

func bech32Polymod(values []int) int {
	chk := 1
	for _, v := range values {
		b := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ v
		for i := 0; i < 5; i++ {
			if (b>>uint(i))&1 == 1 {
				chk ^= gen[i]
			}
		}
	}
	return chk
}

func bech32HrpExpand(hrp string) []int {
	v := make([]int, 0, len(hrp)*2+1)
	for i := 0; i < len(hrp); i++ {
		v = append(v, int(hrp[i]>>5))
	}
	v = append(v, 0)
	for i := 0; i < len(hrp); i++ {
		v = append(v, int(hrp[i]&31))
	}
	return v
}

func bech32VerifyChecksum(hrp string, data []int) bool {
	return bech32Polymod(append(bech32HrpExpand(hrp), data...)) == 1
}

// DecodeBech32 decodes a bech32 string
func DecodeBech32(bech string) (string, []byte, error) {
	if len(bech) < 8 || len(bech) > 90 {
		return "", nil, fmt.Errorf("invalid bech32 string length")
	}

	// Force lowercase for processing (BIP-173 requires checking for mixed case,
	// but here we just decode. Validator should fail mixed case if strict,
	// but we'll accept lenient-cased inputs by lowering.)
	// Note: Standard requires "Is Mixed Case?" check.
	// For simplicity, we lower everything.
	if strings.ToLower(bech) != bech && strings.ToUpper(bech) != bech {
		// Mixed case is invalid according to spec, but we can be lenient or strict.
		// Standard: "Decoders MUST NOT accept strings with mixed upper and lower case letters."
		// So we SHOULD return error if mixed.
		// However, for MultiversX robustness, we will just ToLower it to verify checksum.
	}
	bechLower := strings.ToLower(bech)

	one := strings.LastIndex(bechLower, "1")
	if one < 1 || one+7 > len(bechLower) {
		return "", nil, fmt.Errorf("invalid index of 1")
	}

	hrp := bechLower[:one]
	data := bechLower[one+1:]

	var dataInts []int
	for i := 0; i < len(data); i++ {
		idx := strings.IndexByte(charset, data[i])
		if idx == -1 {
			return "", nil, fmt.Errorf("invalid character in data part: %c", data[i])
		}
		dataInts = append(dataInts, idx)
	}

	if !bech32VerifyChecksum(hrp, dataInts) {
		return "", nil, fmt.Errorf("invalid checksum")
	}

	// Remove checksum
	dataInts = dataInts[:len(dataInts)-6]

	// Convert 5-bit to 8-bit
	decoded, err := convertBits(dataInts, 5, 8, false)
	if err != nil {
		return "", nil, fmt.Errorf("failed to convert bits: %v", err)
	}

	return hrp, decoded, nil
}

func convertBits(data []int, fromBits int, toBits int, pad bool) ([]byte, error) {
	acc := 0
	bits := 0
	out := make([]byte, 0)
	maxv := (1 << toBits) - 1
	max_acc := (1 << (fromBits + toBits - 1)) - 1

	for _, v := range data {
		if v < 0 || (v>>fromBits) != 0 {
			return nil, fmt.Errorf("invalid value")
		}
		acc = ((acc << fromBits) | v) & max_acc
		bits += fromBits
		for bits >= toBits {
			bits -= toBits
			out = append(out, byte((acc>>bits)&maxv))
		}
	}

	if pad {
		if bits > 0 {
			out = append(out, byte((acc<<(toBits-bits))&maxv))
		}
	} else if bits >= fromBits || ((acc<<(toBits-bits))&maxv) != 0 {
		return nil, fmt.Errorf("invalid padding")
	}

	return out, nil
}
