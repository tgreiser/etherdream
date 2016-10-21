package main

import (
	"io"
	"log"
	"runtime"

	"github.com/tgreiser/etherdream"
)

func main() {
	log.Printf("Listening...\n")
	addr, _, err := etherdream.FindFirstDAC()
	if err != nil {
		log.Fatalf("Network error: %v", err)
	}

	log.Printf("Found DAC at %v\n", addr)

	dac := etherdream.NewDAC(addr.IP.String())
	err = dac.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer dac.Close()
	log.Printf("Initialized:  %v\n", dac.LastStatus)

	dac.Play(squarePointStream)
}

func squarePointStream(w *io.PipeWriter) etherdream.Points {
	defer w.Close()
	pmax := 15600
	pstep := 100
	cmax := 65535
	for {
		for x := range xrange(-pmax, pmax, pstep) {
			w.Write(etherdream.NewPoint(x, pmax, cmax, 0, 0, cmax).Encode())
		}
		for y := range xrange(pmax, -pmax, -pstep) {
			w.Write(etherdream.NewPoint(pmax, y, 0, cmax, 0, cmax).Encode())
		}
		for x := range xrange(pmax, -pmax, -pstep) {
			w.Write(etherdream.NewPoint(x, -pmax, 0, 0, cmax, cmax).Encode())
		}
		for y := range xrange(-pmax, pmax, pstep) {
			w.Write(etherdream.NewPoint(-pmax, y, cmax, cmax, cmax, cmax).Encode())
		}
		log.Printf("Generated a frame")
		runtime.Gosched() // yield for other go routines
	}
}

func xrange(min, max, step int) []int {
	ret := make([]int, (max-min)/step+1)
	iY := 0
	for iX := min; iX <= max; iX += step {
		ret[iY] = iX
		iY++
	}
	return ret
}
