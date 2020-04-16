package std

import (
	"fmt"
	"regexp"
	"strings"
)

// NewCompactBitArray returns a new compact bit array.
// It returns nil if the number of bits is zero.
func NewCompactBitArray(bits int) *CompactBitArray {
	if bits <= 0 {
		return nil
	}
	return &CompactBitArray{
		ExtraBitsStored: []byte{byte(bits % 8)},
		Elems:           make([]byte, (bits+7)/8),
	}
}

// Size returns the number of bits in the bitarray
func (bA *CompactBitArray) Size() int {
	if bA == nil {
		return 0
	} else if bA.ExtraBitsStored[0] == byte(0) {
		return len(bA.Elems) * 8
	}
	// num_bits = 8*num_full_bytes + overflow_in_last_byte
	// num_full_bytes = (len(bA.Elems)-1)
	return (len(bA.Elems)-1)*8 + int(bA.ExtraBitsStored[0])
}

// GetIndex returns the bit at index i within the bit array.
// The behavior is undefined if i >= bA.Size()
func (bA *CompactBitArray) GetIndex(i int) bool {
	if bA == nil {
		return false
	}
	if i >= bA.Size() {
		return false
	}
	return bA.Elems[i>>3]&(uint8(1)<<uint8(7-(i%8))) > 0
}

// SetIndex sets the bit at index i within the bit array.
// The behavior is undefined if i >= bA.Size()
func (bA *CompactBitArray) SetIndex(i int, v bool) bool {
	if bA == nil {
		return false
	}
	if i >= bA.Size() {
		return false
	}
	if v {
		bA.Elems[i>>3] |= (uint8(1) << uint8(7-(i%8)))
	} else {
		bA.Elems[i>>3] &= ^(uint8(1) << uint8(7-(i%8)))
	}
	return true
}

// NumTrueBitsBefore returns the number of bits set to true before the
// given index. e.g. if bA = _XX__XX, NumOfTrueBitsBefore(4) = 2, since
// there are two bits set to true before index 4.
func (bA *CompactBitArray) NumTrueBitsBefore(index int) int {
	numTrueValues := 0
	for i := 0; i < index; i++ {
		if bA.GetIndex(i) {
			numTrueValues++
		}
	}
	return numTrueValues
}

// Copy returns a copy of the provided bit array.
func (bA *CompactBitArray) Copy() *CompactBitArray {
	if bA == nil {
		return nil
	}
	c := make([]byte, len(bA.Elems))
	copy(c, bA.Elems)
	return &CompactBitArray{
		ExtraBitsStored: bA.ExtraBitsStored,
		Elems:           c,
	}
}

// String returns a string representation of CompactBitArray: BA{<bit-string>},
// where <bit-string> is a sequence of 'x' (1) and '_' (0).
// The <bit-string> includes spaces and newlines to help people.
// For a simple sequence of 'x' and '_' characters with no spaces or newlines,
// see the MarshalJSON() method.
// Example: "BA{_x_}" or "nil-BitArray" for nil.
func (bA *CompactBitArray) String() string {
	return bA.StringIndented("")
}

// StringIndented returns the same thing as String(), but applies the indent
// at every 10th bit, and twice at every 50th bit.
func (bA *CompactBitArray) StringIndented(indent string) string {
	if bA == nil {
		return "nil-BitArray"
	}
	lines := []string{}
	bits := ""
	size := bA.Size()
	for i := 0; i < size; i++ {
		if bA.GetIndex(i) {
			bits += "x"
		} else {
			bits += "_"
		}
		if i%100 == 99 {
			lines = append(lines, bits)
			bits = ""
		}
		if i%10 == 9 {
			bits += indent
		}
		if i%50 == 49 {
			bits += indent
		}
	}
	if len(bits) > 0 {
		lines = append(lines, bits)
	}
	return fmt.Sprintf("BA{%v:%v}", size, strings.Join(lines, indent))
}

// MarshalJSON implements json.Marshaler interface by marshaling bit array
// using a custom format: a string of '-' or 'x' where 'x' denotes the 1 bit.
func (bA *CompactBitArray) MarshalJSON() ([]byte, error) {
	if bA == nil {
		return []byte("null"), nil
	}

	bits := `"`
	size := bA.Size()
	for i := 0; i < size; i++ {
		if bA.GetIndex(i) {
			bits += `x`
		} else {
			bits += `_`
		}
	}
	bits += `"`
	return []byte(bits), nil
}

var bitArrayJSONRegexp = regexp.MustCompile(`\A"([_x]*)"\z`)

// UnmarshalJSON implements json.Unmarshaler interface by unmarshaling a custom
// JSON description.
func (bA *CompactBitArray) UnmarshalJSON(bz []byte) error {
	b := string(bz)
	if b == "null" {
		// This is required e.g. for encoding/json when decoding
		// into a pointer with pre-allocated BitArray.
		bA.ExtraBitsStored = []byte{0}
		bA.Elems = nil
		return nil
	}

	// Validate 'b'.
	match := bitArrayJSONRegexp.FindStringSubmatch(b)
	if match == nil {
		return fmt.Errorf("bitArray in JSON should be a string of format %q but got %s", bitArrayJSONRegexp.String(), b)
	}
	bits := match[1]

	// Construct new CompactBitArray and copy over.
	numBits := len(bits)
	bA2 := NewCompactBitArray(numBits)
	for i := 0; i < numBits; i++ {
		if bits[i] == 'x' {
			bA2.SetIndex(i, true)
		}
	}
	*bA = *bA2
	return nil
}

// // CompactMarshal is a space efficient encoding for CompactBitArray.
// // It is not amino compatible.
// func (bA *CompactBitArray) Marshal() []byte {
// 	size := bA.Size()
// 	if size <= 0 {
// 		return []byte("null")
// 	}
// 	bz := make([]byte, 0, size/8)
// 	// length prefix number of bits, not number of bytes. This difference
// 	// takes 3-4 bits in encoding, as opposed to instead encoding the number of
// 	// bytes (saving 3-4 bits) and including the offset as a full byte.
// 	bz = appendUvarint(bz, uint64(size))
// 	bz = append(bz, bA.Elems...)
// 	return bz
// }

// // CompactUnmarshal is a space efficient decoding for CompactBitArray.
// // It is not amino compatible.
// func (bA *CompactBitArray) Unmarshal(bz []byte) error {
// 	if len(bz) < 2 {
// 		return errors.New("compact bit array: invalid compact unmarshal size")
// 	} else if bytes.Equal(bz, []byte("null")) {
// 		return nil
// 	}
// 	size, n := binary.Uvarint(bz)
// 	bz = bz[n:]
// 	if len(bz) != int(size+7)/8 {
// 		return errors.New("compact bit array: invalid compact unmarshal size")
// 	}

// 	bA = NewCompactBitArray(int(size % 8))
// 	bA.Elems = bz
// 	return nil
// }

// func appendUvarint(b []byte, x uint64) []byte {
// 	var a [binary.MaxVarintLen64]byte
// 	n := binary.PutUvarint(a[:], x)
// 	return append(b, a[:n]...)
// }
