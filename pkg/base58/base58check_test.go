// Copyright (c) 2013-2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package base58_test

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/olympus-protocol/ogen/pkg/base58"
)

var checkEncodingStringTests = []struct {
	version byte
	in      string
	out     string
}{
	{20, "", "3MNQE1X"},
	{20, " ", "B2Kr6dBE"},
	{20, "-", "B3jv1Aft"},
	{20, "0", "B482yuaX"},
	{20, "1", "B4CmeGAC"},
	{20, "-1", "mM7eUf6kB"},
	{20, "11", "mP7BMTDVH"},
	{20, "abc", "4QiVtDjUdeq"},
	{20, "1234598760", "ZmNb8uQn5zvnUohNCEPP"},
	{20, "abcdefghijklmnopqrstuvwxyz", "K2RYDcKfupxwXdWhSAxQPCeiULntKm63UXyx5MvEH2"},
	{20, "00000000000000000000000000000000000000000000000000000000000000", "bi1EWXwJay2udZVxLJozuTb8Meg4W9c6xnmJaRDjg6pri5MBAxb9XwrpQXbtnqEoRV5U2pixnFfwyXC8tRAVC8XxnjK"},
}

func TestBase58Check(t *testing.T) {
	for _, test := range checkEncodingStringTests {
		// test encoding
		res := base58.CheckEncode([]byte(test.in), test.version)
		assert.Equal(t, res, test.out)

		// test decoding
		decodeRes, version, err := base58.CheckDecode(test.out)
		assert.NoError(t, err)
		assert.Equal(t, test.version, version)
		assert.Equal(t, test.in, string(decodeRes))

	}

	// test the two decoding failure cases
	// case 1: checksum error
	_, _, err := base58.CheckDecode("3MNQE1Y")
	assert.Equal(t, base58.ErrChecksum, err)

	// case 2: invalid formats (string lengths below 5 mean the version byte and/or the checksum
	// bytes are missing).
	testString := ""
	for l := 0; l < 4; l++ {
		testString = testString + "x"
		_, _, err = base58.CheckDecode(testString)
		assert.Equal(t, base58.ErrInvalidFormat, err)
	}

}
