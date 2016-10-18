package main

import (
	"github.com/tgreiser/etherdream"
	"log"
)

func main() {
	log.Printf("Listening...\n")
	addr, bp, err := etherdream.FindFirstDAC()
	if err != nil {
		log.Fatal("Network error: %v", err)
	}

	log.Printf("Found DAC at %v\n", addr)

	log.Printf("BP:\n%v\n", bp)
	log.Printf("Status:\n%v\n", bp.Status)

	//dac := etherdream.NewDAC(addr)
	//dac.play(square_point_stream())
}
