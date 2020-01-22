// Copyright (c) 2013, 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package amount

import (
	"errors"
	"math"
	"strconv"
)

type AmountUnit int

const (
	AmountMega  AmountUnit = 6
	AmountKilo  AmountUnit = 3
	Amount      AmountUnit = 0
	AmountMilli AmountUnit = -3
	AmountMicro AmountUnit = -6
	AmountSats  AmountUnit = -8

	SatsPerUnit = 1e8
	MaxSats     = 21e6 * SatsPerUnit
)

type AmountType int64

func round(f float64) AmountType {
	if f < 0 {
		return AmountType(f - 0.5)
	}
	return AmountType(f + 0.5)
}

func NewAmount(f float64) (AmountType, error) {
	switch {
	case math.IsNaN(f):
		fallthrough
	case math.IsInf(f, 1):
		fallthrough
	case math.IsInf(f, -1):
		return 0, errors.New("invalid amount")
	}

	return round(f * SatsPerUnit), nil
}

func (a AmountType) ToUnit(u AmountUnit) float64 {
	return float64(a) / math.Pow10(int(u+8))
}

func (a AmountType) ToNormalUnit() float64 {
	return a.ToUnit(Amount)
}

func (a AmountType) Format(u AmountUnit) string {
	return strconv.FormatFloat(a.ToUnit(u), 'f', -int(u+8), 64)
}

func (a AmountType) String() string {
	return a.Format(Amount)
}

func (a AmountType) MulF64(f float64) AmountType {
	return round(float64(a) * f)
}
