package main

import (
	"fmt"
)

func main() {
	const a = 8
	const b = 1<<a - 1

	fmt.Printf("%x", b)
}
