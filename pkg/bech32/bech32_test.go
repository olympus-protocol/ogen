// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bech32_test

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"testing"

	"github.com/olympus-protocol/ogen/pkg/bech32"
)

func TestBech32(t *testing.T) {
	tests := []struct {
		str   string
		valid bool
	}{
		{"A12UEL5L", true},
		{"an83characterlonghumanreadablepartthatcontainsthenumber1andtheexcludedcharactersbio1tt5tgs", true},
		{"abcdef1qpzry9x8gf2tvdw0s3jn54khce6mua7lmqqqxw", true},
		{"11qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqc8247j", true},
		{"split1checkupstagehandshakeupstreamerranterredcaperred2y9e3w", true},
		{"split1checkupstagehandshakeupstreamerranterredcaperred2y9e2w", false},                         // invalid checksum
		{"s lit1checkupstagehandshakeupstreamerranterredcaperredp8hs2p", false},                         // invalid character (space) in hrp
		{"spl" + strconv.Itoa(127) + "t1checkupstagehandshakeupstreamerranterredcaperred2y9e3w", false}, // invalid character (DEL) in hrp
		{"split1cheo2y9e2w", false}, // invalid character (o) in data part
		{"split1a2y9w", false},      // too short data part
		{"1checkupstagehandshakeupstreamerranterredcaperred2y9e3w", false},                                     // empty hrp
		{"11qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqc8247j", false}, // too long
	}

	for _, test := range tests {
		str := test.str
		hrp, decoded, err := bech32.Decode(str)
		if !test.valid {
			assert.NotNil(t, err)
			continue
		}

		// Valid string decoding should result in no error.
		assert.NoError(t, err)

		// Check that it encodes to the same string
		encoded, err := bech32.Encode(hrp, decoded)
		assert.NoError(t, err)

		assert.Equal(t, strings.ToLower(str), encoded)

		// Flip a bit in the string an make sure it is caught.
		pos := strings.LastIndexAny(str, "1")
		flipped := str[:pos+1] + string(str[pos+1]^1) + str[pos+2:]
		_, _, err = bech32.Decode(flipped)
		assert.NotNil(t, err)

	}
}
