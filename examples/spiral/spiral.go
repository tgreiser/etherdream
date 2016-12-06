/*
# Copyright 2016 Tim Greiser
# Based on work by Jacob Potter, some comments are from his
# protocol documents. Example code from Brandon Thomas.
#
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
	"github.com/tgreiser/ln/ln"
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

	debug := false
	dac.Play(pointStream, debug)
}

func pointStream(w io.WriteCloser) {
	defer w.Close()

	c := color.RGBA{0x88, 0x00, 0x55, 0xFF}
	maxrad := 22600
	rad := int(maxrad / 100)
	var spiralgrowth float64 = 14
	frame := 0

	for {
		var j float64

		for _, i := range xrange(0, 1000, 1) {
			f := float64(i) / 1000.0 * 2.0 * math.Pi * spiralgrowth
			j = f
			x := int(j * math.Cos(f) * float64(rad))
			y := int(j * math.Sin(f) * float64(rad))
			w.Write(etherdream.NewPoint(x, y, c).Encode())
		}

		// blank and return to origin
		f := 1000.0 / 1000.0 * 2.0 * math.Pi * spiralgrowth
		j = f
		x := j * math.Cos(f) * float64(rad)
		y := j * math.Sin(f) * float64(rad)
		p := ln.Path{ln.Vector{x, y, 0}, ln.Vector{0, 0, 0}}
		etherdream.BlankPath(w, p)

		frame++
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
