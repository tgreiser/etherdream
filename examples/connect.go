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

	dac := etherdream.NewDAC(addr.IP.String())
	err = dac.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer dac.Close()
	log.Printf("Initialized %v\n", d.LastStatus)
	log.Printf("Firmware String: %v\n", d.FirmwareString)

	st, err := dac.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Ping status: %v", st)
}
