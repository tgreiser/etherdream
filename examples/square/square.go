/*
# Copyright 2016 Tim Greiser
# Based on work by Jacob Potter, some comments are from his
# protocol documents. Example code from Brandon Thomas.

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
	"math"

	"image/color"

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

	dac.Play(squarePointStream)
}

func squareFrame(w io.WriteCloser, pmax, pstep int) *etherdream.Point {
	for _, x := range xrange(-pmax, pmax, pstep) {
		w.Write(etherdream.NewPoint(x, pmax, color.RGBA{0xff, 0x00, 0x00, 0xff}).Encode())
	}
	for _, y := range xrange(pmax, -pmax, -pstep) {
		w.Write(etherdream.NewPoint(pmax, y, color.RGBA{0x00, 0xff, 0x00, 0xff}).Encode())
	}
	for _, x := range xrange(pmax, -pmax, -pstep) {
		w.Write(etherdream.NewPoint(x, -pmax, color.RGBA{0x00, 0x00, 0xff, 0xff}).Encode())
	}
	var pt *etherdream.Point
	for _, y := range xrange(-pmax, pmax, pstep) {
		pt = etherdream.NewPoint(-pmax, y, color.RGBA{0xff, 0xff, 0xff, 0xff})
		w.Write(pt.Encode())
	}
	return pt
}

func squarePointStream(w io.WriteCloser) {
	defer w.Close()
	pmax := 5600
	pstep := 112
	pct := pmax / pstep * 4
	for {
		var pt *etherdream.Point
		ct := pct
		times := int(math.Floor(float64(etherdream.FramePoints() / ct)))
		ct = 0

		// This approach gives flicker free draw, when
		// repeated 4x
		for i := 0; i < times; i++ {
			pt = squareFrame(w, pmax, pstep)
			ct += pct
		}

		etherdream.NextFrame(w, ct, *pt)
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
