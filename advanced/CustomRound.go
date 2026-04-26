package main

import (
	"fmt"
)

func main() {
	fmt.Println(customRound(4.2))  // 4.0
	fmt.Println(customRound(4.3))  // 4.5
	fmt.Println(customRound(4.7))  // 5.0
	fmt.Println(customRound(10.8)) // 10.0
	fmt.Println(customRound(10.6)) // 10.5
}

func customRound(num float64) (round float64) {
	intPart := int64(num)
	frac := int64(num*100) % 100
	// frac := num - float64(intPart)
	// frac = float64(int((frac+1e-9)*10)) / 10
	// fmt.Println("frac:", frac)
	if frac <= 20 {
		round = float64(intPart)
	} else if frac <= 70 {
		round = float64(intPart) + 0.5
	} else {
		round = float64(intPart + 1)
	}
	return
}
