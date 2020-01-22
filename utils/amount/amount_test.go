package amount

import (
	"math"
	"testing"
)

func TestAmountCreation(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		valid    bool
		expected AmountType
	}{
		// Positive tests.
		{
			name:     "zero",
			amount:   0,
			valid:    true,
			expected: 0,
		},
		{
			name:     "max producible",
			amount:   21e6,
			valid:    true,
			expected: MaxSats,
		},
		{
			name:     "min producible",
			amount:   -21e6,
			valid:    true,
			expected: -MaxSats,
		},
		{
			name:     "exceeds max producible",
			amount:   21e6 + 1e-8,
			valid:    true,
			expected: MaxSats + 1,
		},
		{
			name:     "exceeds min producible",
			amount:   -21e6 - 1e-8,
			valid:    true,
			expected: -MaxSats - 1,
		},
		{
			name:     "one hundred",
			amount:   100,
			valid:    true,
			expected: 100 * SatsPerUnit,
		},
		{
			name:     "fraction",
			amount:   0.01234567,
			valid:    true,
			expected: 1234567,
		},
		{
			name:     "rounding up",
			amount:   54.999999999999943157,
			valid:    true,
			expected: 55 * SatsPerUnit,
		},
		{
			name:     "rounding down",
			amount:   55.000000000000056843,
			valid:    true,
			expected: 55 * SatsPerUnit,
		},

		// Negative tests.
		{
			name:   "not-a-number",
			amount: math.NaN(),
			valid:  false,
		},
		{
			name:   "-infinity",
			amount: math.Inf(-1),
			valid:  false,
		},
		{
			name:   "+infinity",
			amount: math.Inf(1),
			valid:  false,
		},
	}

	for _, test := range tests {
		a, err := NewAmount(test.amount)
		switch {
		case test.valid && err != nil:
			t.Errorf("%v: Positive test Amount creation failed with: %v", test.name, err)
			continue
		case !test.valid && err == nil:
			t.Errorf("%v: Negative test Amount creation succeeded (value %v) when should fail", test.name, a)
			continue
		}

		if a != test.expected {
			t.Errorf("%v: Created amount %v does not match expected %v", test.name, a, test.expected)
			continue
		}
	}
}

func TestAmountUnitConversions(t *testing.T) {
	tests := []struct {
		amount    AmountType
		unit      AmountUnit
		converted float64
		s         string
	}{
		{
			amount:    MaxSats,
			unit:      AmountMega,
			converted: 21,
			s:         "21",
		},
		{
			amount:    44433322211100,
			unit:      AmountKilo,
			converted: 444.33322211100,
			s:         "444.333222111",
		},
		{
			amount:    44433322211100,
			unit:      Amount,
			converted: 444333.22211100,
			s:         "444333.222111",
		},
		{
			amount:    44433322211100,
			unit:      AmountMilli,
			converted: 444333222.11100,
			s:         "444333222.111",
		},
		{
			amount:    44433322211100,
			unit:      AmountMicro,
			converted: 444333222111.00,
			s:         "444333222111",
		},
		{
			amount:    44433322211100,
			unit:      AmountSats,
			converted: 44433322211100,
			s:         "44433322211100",
		},
		{
			amount:    44433322211100,
			unit:      AmountUnit(-1),
			converted: 4443332.2211100,
			s:         "4443332.22111",
		},
	}

	for _, test := range tests {
		f := test.amount.ToUnit(test.unit)
		if f != test.converted {
			t.Errorf("%v: converted value %v does not match expected %v", test.s, f, test.converted)
			continue
		}

		s := test.amount.Format(test.unit)
		if s != test.s {
			t.Errorf("%v: format '%v' does not match expected '%v'", test.s, s, test.s)
			continue
		}

		// Verify that AmountType.ToNormalUnit works as advertised.
		f1 := test.amount.ToUnit(Amount)
		f2 := test.amount.ToNormalUnit()
		if f1 != f2 {
			t.Errorf("%v: ToNormalUnit does not match ToUnit(Amount): %v != %v", test.s, f1, f2)
		}

		// Verify that AmountType.String works as advertised.
		s1 := test.amount.Format(Amount)
		s2 := test.amount.String()
		if s1 != s2 {
			t.Errorf("%v: String does not match Format(Amount): %v != %v", test.s, s1, s2)
		}
	}
}

func TestAmountMulF64(t *testing.T) {
	tests := []struct {
		name string
		amt  AmountType
		mul  float64
		res  AmountType
	}{
		{
			name: "Multiply 0.1 by 2",
			amt:  100e5, // 0.1
			mul:  2,
			res:  200e5, // 0.2
		},
		{
			name: "Multiply 0.2 by 0.02",
			amt:  200e5, // 0.2
			mul:  1.02,
			res:  204e5, // 0.204
		},
		{
			name: "Multiply 0.1 by -2",
			amt:  100e5, // 0.1
			mul:  -2,
			res:  -200e5, // -0.2
		},
		{
			name: "Multiply 0.2  by -0.02",
			amt:  200e5, // 0.2
			mul:  -1.02,
			res:  -204e5, // -0.204
		},
		{
			name: "Multiply -0.1 by 2",
			amt:  -100e5, // -0.1
			mul:  2,
			res:  -200e5, // -0.2
		},
		{
			name: "Multiply -0.2 by 0.02",
			amt:  -200e5, // -0.2
			mul:  1.02,
			res:  -204e5, // -0.204
		},
		{
			name: "Multiply -0.1 by -2",
			amt:  -100e5, // -0.1
			mul:  -2,
			res:  200e5, // 0.2
		},
		{
			name: "Multiply -0.2 by -0.02",
			amt:  -200e5, // -0.2
			mul:  -1.02,
			res:  204e5, // 0.204
		},
		{
			name: "Round down",
			amt:  49, // 49
			mul:  0.01,
			res:  0,
		},
		{
			name: "Round up",
			amt:  50, // 50
			mul:  0.01,
			res:  1, // 1
		},
		{
			name: "Multiply by 0.",
			amt:  1e8, // 1
			mul:  0,
			res:  0, // 0
		},
		{
			name: "Multiply 1 by 0.5.",
			amt:  1, // 1
			mul:  0.5,
			res:  1, // 1
		},
		{
			name: "Multiply 100 by 66%.",
			amt:  100, // 100
			mul:  0.66,
			res:  66, // 66
		},
		{
			name: "Multiply 100 by 66.6%.",
			amt:  100, // 100
			mul:  0.666,
			res:  67, // 67
		},
		{
			name: "Multiply 100 by 2/3.",
			amt:  100, // 100
			mul:  2.0 / 3,
			res:  67, // 67
		},
	}

	for _, test := range tests {
		a := test.amt.MulF64(test.mul)
		if a != test.res {
			t.Errorf("%v: expected %v got %v", test.name, test.res, a)
		}
	}
}
