package domain

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

var decimalPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]{1,18})?$`)

// Decimal keeps monetary comparisons, addition and subtraction exact.
type Decimal struct {
	coefficient *big.Int
	scale       int
}

func ParseDecimal(s string) (Decimal, error) {
	if len(s) > 79 || !decimalPattern.MatchString(s) {
		return Decimal{}, fmt.Errorf("use a nonnegative decimal with at most 78 digits and 18 decimal places")
	}
	parts := strings.Split(s, ".")
	digits, scale := parts[0], 0
	if len(parts) == 2 {
		digits += parts[1]
		scale = len(parts[1])
	}
	if len(digits) > 78 {
		return Decimal{}, fmt.Errorf("amount exceeds 78 digits")
	}
	n, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return Decimal{}, fmt.Errorf("invalid decimal")
	}
	return Decimal{n, scale}, nil
}

func powerOfTen(scale int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
}

func (d Decimal) Compare(other Decimal) int {
	a, b := new(big.Int).Set(d.coefficient), new(big.Int).Set(other.coefficient)
	if d.scale < other.scale {
		a.Mul(a, powerOfTen(other.scale-d.scale))
	} else if other.scale < d.scale {
		b.Mul(b, powerOfTen(d.scale-other.scale))
	}
	return a.Cmp(b)
}

func (d Decimal) Add(other Decimal) Decimal {
	a, b, scale := d.align(other)
	return Decimal{new(big.Int).Add(a, b), scale}
}

func (d Decimal) Subtract(other Decimal) Decimal {
	a, b, scale := d.align(other)
	return Decimal{new(big.Int).Sub(a, b), scale}
}

func (d Decimal) align(other Decimal) (*big.Int, *big.Int, int) {
	a, b := new(big.Int).Set(d.coefficient), new(big.Int).Set(other.coefficient)
	scale := d.scale
	if scale < other.scale {
		scale = other.scale
	}
	a.Mul(a, powerOfTen(scale-d.scale))
	b.Mul(b, powerOfTen(scale-other.scale))
	return a, b, scale
}

func ParseMoney(value string) (Decimal, error) {
	amount, err := ParseDecimal(value)
	if err != nil {
		return Decimal{}, err
	}
	parts := strings.Split(amount.String(), ".")
	if len(parts) == 2 && len(parts[1]) > 2 {
		return Decimal{}, fmt.Errorf("use at most 2 decimal places")
	}
	return amount, nil
}

func (d Decimal) String() string {
	s := new(big.Int).Abs(d.coefficient).String()
	sign := ""
	if d.coefficient.Sign() < 0 {
		sign = "-"
	}
	if d.scale == 0 {
		return sign + s
	}
	if len(s) <= d.scale {
		s = strings.Repeat("0", d.scale-len(s)+1) + s
	}
	s = s[:len(s)-d.scale] + "." + s[len(s)-d.scale:]
	return sign + strings.TrimRight(strings.TrimRight(s, "0"), ".")
}
