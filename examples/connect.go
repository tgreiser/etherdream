package main

import (
	"fmt"
	"github.com/tgreiser/etherdream"
)

func main() {
	fmt.Printf("Listening...\n")
	addr, bp, err := etherdream.FindFirstDAC()

	fmt.Printf("Found DAC at %v\nBroadcast Packet %v\nerr %v", addr, bp, err)
}
