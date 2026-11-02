// 运行示例：PRJ-02
package main

import (
	"fmt"
	"math/big"
)

func main() {
	a, ok := new(big.Rat).SetString("60000.10")
	if !ok {
		panic("bad a")
	}
	b, ok := new(big.Rat).SetString("0.20")
	if !ok {
		panic("bad b")
	}
	total := new(big.Rat).Add(a, b)
	fmt.Println(total.FloatString(2))
	fmt.Println(total.Cmp(a) > 0)
}
