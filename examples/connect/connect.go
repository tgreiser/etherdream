package main

import (
	"log"

	"github.com/tgreiser/etherdream"
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

	dac := etherdream.NewDAC(addr.IP.String())
	err = dac.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer dac.Close()
	log.Printf("Initialized %v\n", dac.LastStatus)
	log.Printf("Firmware String: %v\n", dac.FirmwareString)

	st, err := dac.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Ping status: %v", st)
}
