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
	"runtime"

	"math"

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

	debug := false
	dac.Play(pointStream, debug)
}

func pointStream(w *io.PipeWriter) etherdream.Points {
	defer w.Close()

	cmax := etherdream.ScaleColor(.5)
	maxrad := 22600
	rad := int(maxrad / 100)
	var spiralgrowth float64 = 14
	blankPts := 25
	frame := 0

	for {
		var j float64

		for _, i := range xrange(0, 1000, 1) {
			f := float64(i) / 1000.0 * 2.0 * math.Pi * spiralgrowth
			j = f
			x := int(j * math.Cos(f) * float64(rad))
			y := int(j * math.Sin(f) * float64(rad))
			w.Write(etherdream.NewPoint(x, y, cmax, cmax, cmax, cmax).Encode())
		}

		// blank and return to origin
		f := 1000.0 / 1000.0 * 2.0 * math.Pi * spiralgrowth
		j = f
		x := int(j * math.Cos(f) * float64(rad))
		y := int(j * math.Sin(f) * float64(rad))

		// first lets throw in a few without moving
		for i := 1; i <= 25; i++ {
			w.Write(etherdream.NewPoint(x-x/blankPts, y-y/blankPts, 0, 0, 0, 0).Encode())
		}

		// move back to origin
		for i := 1; i <= blankPts; i++ {
			//log.Printf("x %v y %v\n", x-x*i/blankPts, y-y*i/blankPts)
			w.Write(etherdream.NewPoint(x-x*i/blankPts, y-y*i/blankPts, 0, 0, 0, 0).Encode())
		}

		// few more still points
		for i := 1; i <= 25; i++ {
			w.Write(etherdream.NewPoint(x-x*blankPts/blankPts, y-y*blankPts/blankPts, 0, 0, 0, 0).Encode())
		}

		frame++
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
