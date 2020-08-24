package bech32

import (
	"fmt"
	"strings"
)

// charset is the sequence of ascii characters that make up the bech32
// alphabet.  Each character represents a 5-bit squashed byte.
// q = 0b00000, p = 0b00001, z = 0b00010, and so on.
const charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

// inverseCharset is a mapping of 8-bit ascii characters to the charset
// positions.  Both uppercase and lowercase ascii are mapped to the 5-bit
// position values.
var inverseCharset = [256]int8{
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	15, -1, 10, 17, 21, 20, 26, 30, 7, 5, -1, -1, -1, -1, -1, -1,
	-1, 29, -1, 24, 13, 25, 9, 8, 23, -1, 18, 22, 31, 27, 19, -1,
	1, 0, 3, 16, 11, 28, 12, 14, 6, 4, 2, -1, -1, -1, -1, -1,
	-1, 29, -1, 24, 13, 25, 9, 8, 23, -1, 18, 22, 31, 27, 19, -1,
	1, 0, 3, 16, 11, 28, 12, 14, 6, 4, 2, -1, -1, -1, -1, -1}

// Bytes8to5 extends a byte slice into a longer, padded byte slice of 5-bit elements
// where the high 3 bits are all 0.
func Bytes8to5(input []byte) []byte {
	// no way to triger an error going from 8 to 5
	output, _ := ByteSquasher(input, 8, 5)
	return output
}

// Bytes5to8 goes from squashed bytes to full height bytes
func Bytes5to8(input []byte) ([]byte, error) {
	return ByteSquasher(input, 5, 8)
}

// ByteSquasher squashes full-width (8-bit) bytes into "squashed" 5-bit bytes,
// and vice versa.  It can operate on other widths but in this package only
// goes 5 to 8 and back again.  It can return an error if the squashed input
// you give it isn't actually squashed, or if there is padding (trailing q characters)
// when going from 5 to 8
func ByteSquasher(input []byte, inputWidth, outputWidth uint32) ([]byte, error) {
	var bitstash, accumulator uint32
	var output []byte
	maxOutputValue := uint32((1 << outputWidth) - 1)
	for i, c := range input {
		if c>>inputWidth != 0 {
			return nil, fmt.Errorf("byte %d (%x) high bits set", i, c)
		}
		accumulator = (accumulator << inputWidth) | uint32(c)
		bitstash += inputWidth
		for bitstash >= outputWidth {
			bitstash -= outputWidth
			output = append(output,
				byte((accumulator>>bitstash)&maxOutputValue))
		}
	}
	// pad if going from 8 to 5
	if inputWidth == 8 && outputWidth == 5 {
		if bitstash != 0 {
			output = append(output,
				byte((accumulator << (outputWidth - bitstash) & maxOutputValue)))
		}
	} else if bitstash >= inputWidth ||
		((accumulator<<(outputWidth-bitstash))&maxOutputValue) != 0 {
		// no pad from 5 to 8 allowed
		return nil, fmt.Errorf(
			"invalid padding from %d to %d bits", inputWidth, outputWidth)
	}
	return output, nil
}

// SquashedBytesToString swaps 5-bit bytes with a string of the corresponding letters
func SquashedBytesToString(input []byte) (string, error) {
	var s string
	for i, c := range input {
		if c&0xe0 != 0 {
			return "", fmt.Errorf("high bits set at position %d: %x", i, c)
		}
		s += string(charset[c])
	}
	return s, nil
}

// StringToSquashedBytes uses the inverseCharset to switch from the characters
// back to 5-bit squashed bytes.
func StringToSquashedBytes(input string) ([]byte, error) {
	b := make([]byte, len(input))
	for i, c := range input {
		if inverseCharset[c] == -1 {
			return nil, fmt.Errorf("contains invalid character %s", string(c))
		}
		b[i] = byte(inverseCharset[c])
	}
	return b, nil
}

// PolyMod takes a byte slice and returns the 32-bit BCH checksum.
// Note that the input bytes to PolyMod need to be squashed to 5-bits tall
// before being used in this function.  And this function will not error,
// but instead return an unusable checksum, if you give it full-height bytes.
func PolyMod(values []byte) uint32 {

	// magic generator uint32s
	gen := []uint32{
		0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3,
	}

	// start with 1
	chk := uint32(1)

	for _, v := range values {
		top := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ uint32(v)
		for i, g := range gen {
			if (top>>uint8(i))&1 == 1 {
				chk ^= g
			}
		}
	}

	return chk
}

// HRPExpand turns the human redable part into 5bit-bytes for later processing
func HRPExpand(input string) []byte {
	output := make([]byte, (len(input)*2)+1)

	// first half is the input string shifted down 5 bits.
	// not much is going on there in terms of data / entropy
	for i, c := range input {
		output[i] = uint8(c) >> 5
	}
	// then there's a 0 byte separator
	// don't need to set 0 byte in the middle, as it starts out that way

	// second half is the input string, with the top 3 bits zeroed.
	// most of the data / entropy will live here.
	for i, c := range input {
		output[i+len(input)+1] = uint8(c) & 0x1f
	}
	return output
}

// create checksum makes a 6-shortbyte checksum from the HRP and data parts
func CreateChecksum(hrp string, data []byte) []byte {
	values := append(HRPExpand(hrp), data...)
	// put 6 zero bytes on at the end
	values = append(values, make([]byte, 6)...)
	//get checksum for whole slice

	// flip the LSB of the checksum data after creating it
	checksum := PolyMod(values) ^ 1

	for i := 0; i < 6; i++ {
		// note that this is NOT the same as converting 8 to 5
		// this is it's own expansion to 6 bytes from 4, chopping
		// off the MSBs.
		values[(len(values)-6)+i] = byte(checksum>>(5*(5-uint32(i)))) & 0x1f
	}

	return values[len(values)-6:]
}

func VerifyChecksum(hrp string, data []byte) bool {
	values := append(HRPExpand(hrp), data...)
	checksum := PolyMod(values)
	// make sure it's 1 (from the LSB flip in CreateChecksum
	return checksum == 1
}

// Encode takes regular bytes of data, and an hrp prefix, and returns the
// bech32 encoded string.
func Encode(hrp string, data []byte) string {
	fiveData := Bytes8to5(data)
	return EncodeSquashed(hrp, fiveData)
}

// EncodeSquashed takes the hrp prefix, as well as byte data that has already
// been squashed to 5-bits high, and returns the bech32 encoded string.
// It does not return an error; if you give it non-squashed data it will return
// an empty string.
func EncodeSquashed(hrp string, data []byte) string {
	combined := append(data, CreateChecksum(hrp, data)...)

	// Should be squashed, return empty string if it's not.
	dataString, err := SquashedBytesToString(combined)
	if err != nil {
		return ""
	}
	return hrp + "1" + dataString
}

// Decode takes a bech32 encoded string and returns the hrp and the full-height
// data.  Can error out for various reasons, mostly problems in the string given.
func Decode(adr string) (string, []byte, error) {
	hrp, squashedData, err := DecodeSquashed(adr)
	if err != nil {
		return hrp, nil, err
	}
	data, err := Bytes5to8(squashedData)
	if err != nil {
		return hrp, nil, err
	}
	return hrp, data, nil
}

// DecodeSquashed is the same as Decode, but will return squashed 5-bit high
// data.
func DecodeSquashed(adr string) (string, []byte, error) {

	// make an all lowercase and all uppercase version of the input string
	lowAdr := strings.ToLower(adr)
	highAdr := strings.ToUpper(adr)

	// if there's mixed case, that's not OK
	if adr != lowAdr && adr != highAdr {
		return "", nil, fmt.Errorf("mixed case address")
	}

	// default to lowercase
	adr = lowAdr

	// find the last "1" and split there
	splitLoc := strings.LastIndex(adr, "1")
	if splitLoc == -1 {
		return "", nil, fmt.Errorf("1 separator not present in address")
	}

	// hrp comes before the split
	hrp := adr[0:splitLoc]

	// get squashed data
	data, err := StringToSquashedBytes(adr[splitLoc+1:])
	if err != nil {
		return hrp, nil, err
	}

	// make sure checksum works
	sumOK := VerifyChecksum(hrp, data)
	if !sumOK {
		return hrp, nil, fmt.Errorf("Checksum invalid")
	}

	// chop off checksum to return only payload
	data = data[:len(data)-6]

	return hrp, data, nil
}