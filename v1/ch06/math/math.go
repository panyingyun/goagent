package main

import (
	"fmt"
	"math/big"
	"os"
	"strconv"
)

// 阶乘（big.Int）
func factorial(n int64) *big.Int {
	if n <= 1 {
		return big.NewInt(1)
	}
	res := big.NewInt(1)
	for i := int64(2); i <= n; i++ {
		res.Mul(res, big.NewInt(i))
	}
	return res
}

func main() {
	// 默认计算 1000 位小数，也可通过命令行参数指定
	digits := 1000
	if len(os.Args) > 1 {
		if d, err := strconv.Atoi(os.Args[1]); err == nil && d > 0 {
			digits = d
		}
	}

	// 二进制精度，留出足够余量保证舍入正确
	prec := uint(float64(digits)*3.321928094887362) + 64

	// 拉马努金级数：
	// 1/π = (2√2 / 9801) * Σ_{k=0}^∞ (4k)! * (1103+26390k) / (k!^4 * 396^(4k))
	// 因此 π = 9801 / (2√2 * Σ)
	sum := new(big.Rat)

	for k := int64(0); ; k++ {
		// 计算每一项：term = (4k)! * (1103+26390k) / (k!^4 * 396^(4k))
		fourK := 4 * k

		fact4K := factorial(fourK)
		factK := factorial(k)
		factK4 := new(big.Int).Exp(factK, big.NewInt(4), nil)
		pow396 := new(big.Int).Exp(big.NewInt(396), big.NewInt(fourK), nil)

		linear := 1103 + 26390*k
		numerator := new(big.Int).Mul(fact4K, big.NewInt(linear))
		denominator := new(big.Int).Mul(factK4, pow396)

		term := new(big.Rat).SetFrac(numerator, denominator)
		sum.Add(sum, term)

		// 当项小于 10^-digits 时停止
		if k > 0 {
			threshold := new(big.Rat).SetFrac(
				big.NewInt(1),
				new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil),
			)
			if term.Cmp(threshold) < 0 {
				break
			}
		}
	}

	// 计算 π = 9801 / (2√2 * sum)
	sqrt2 := new(big.Float).SetPrec(prec).Sqrt(big.NewFloat(2))
	sumFloat := new(big.Float).SetPrec(prec).SetRat(sum)

	denominator := new(big.Float).SetPrec(prec).Mul(sqrt2, big.NewFloat(2))
	denominator.Mul(denominator, sumFloat)

	pi := new(big.Float).SetPrec(prec).Quo(big.NewFloat(9801), denominator)

	// 输出指定小数位数的 π
	fmt.Println(pi.Text('f', digits))
}
