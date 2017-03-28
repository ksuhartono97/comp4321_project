package main

import (
	"fmt"
	"math"
)

func Sqrt(x float64) float64 {
	z := float64(1)
	prevz := float64(1);
	count:= 1

	z = z - (((z*z) - x) / 2*z)
	delta := math.Abs(prevz - z);
	for ; delta > 0.0001; delta = math.Abs(prevz - z){
		prevz = z
		z = z - (((z*z) - x) / 2*z)
    count++
	}
  fmt.Println("Iterations: ", count)
	return z
}

func main() {
	fmt.Println(Sqrt(2))
}
