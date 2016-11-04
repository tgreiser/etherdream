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

	dac, err := etherdream.NewDAC(addr.IP.String())
	if err != nil {
		log.Fatal(err)
	}
	defer dac.Close()
	log.Printf("Initialized:  %v\n", dac.LastStatus)
	log.Printf("Firmware String: %v\n", dac.FirmwareString)

	debug := false
	dac.Play(squarePointStream, debug)
}

func squarePointStream(dac *etherdream.DAC) {
	defer dac.Writer.Close()
	pmax := 15600
	pstep := 100
	cmax := 65535
	for {
		for _, x := range xrange(-pmax, pmax, pstep) {
			dac.WritePoint(etherdream.NewPoint(x, pmax, cmax, 0, 0, cmax))
		}
		for _, y := range xrange(pmax, -pmax, -pstep) {
			dac.WritePoint(etherdream.NewPoint(pmax, y, 0, cmax, 0, cmax))
		}
		for _, x := range xrange(pmax, -pmax, -pstep) {
			dac.WritePoint(etherdream.NewPoint(x, -pmax, 0, 0, cmax, cmax))
		}
		for _, y := range xrange(-pmax, pmax, pstep) {
			dac.WritePoint(etherdream.NewPoint(-pmax, y, cmax, cmax, cmax, cmax))
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
