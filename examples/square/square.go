/*
# Copyright 2016 Tim Greiser
# Based on work by Jacob Potter, some comments are from his
# protocol documents

# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, version 3.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"io"
	"log"
	"runtime"

	"github.com/tgreiser/etherdream"
)

func main() {
	log.Printf("Listening...\n")
	addr, bp, err := etherdream.FindFirstDAC()
	if err != nil {
		log.Fatalf("Network error: %v", err)
	}

	log.Printf("Found DAC at %v\n", addr)
	log.Printf("BP: %v\n\n", bp)

	dac, err := etherdream.NewDAC(addr.IP.String())
	if err != nil {
		log.Fatal(err)
	}
	defer dac.Close()
	log.Printf("Initialized:  %v\n\n", dac.LastStatus)
	log.Printf("Firmware String: %v\n\n", dac.FirmwareString)

	//dac.Measure(squarePointStream)
	debug := false
	dac.Play(squarePointStream, debug)
}

func squarePointStream(w *io.PipeWriter) etherdream.Points {
	defer w.Close()
	pmax := 15600
	pstep := 100
	cmax := 25000 //65535
	for {
		for _, x := range xrange(-pmax, pmax, pstep) {
			w.Write(etherdream.NewPoint(x, pmax, cmax, 0, 0, cmax).Encode())
		}
		for _, y := range xrange(pmax, -pmax, -pstep) {
			w.Write(etherdream.NewPoint(pmax, y, 0, cmax, 0, cmax).Encode())
		}
		for _, x := range xrange(pmax, -pmax, -pstep) {
			w.Write(etherdream.NewPoint(x, -pmax, 0, 0, cmax, cmax).Encode())
		}
		for _, y := range xrange(-pmax, pmax, pstep) {
			w.Write(etherdream.NewPoint(-pmax, y, cmax, cmax, cmax, cmax).Encode())
		}
		//log.Printf("Generated a frame")
		runtime.Gosched() // yield for other go routines
	}
}

func xrange(min, max, step int) []int {
	rng := max - min
	ret := make([]int, rng/step+1)
	iY := 0
	for iX := min; rlogic(min, max, iX); iX += step {
		ret[iY] = iX
		iY++
	}
	return ret
}

func rlogic(min, max, iX int) bool {
	if min < max {
		return iX <= max
	}
	return iX >= max
}
