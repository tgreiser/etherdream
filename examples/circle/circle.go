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
	"runtime"

	"math"

	"image/color"

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

func pointStream(w *io.PipeWriter) {
	defer w.Close()

	pstep := 100 // 30 and below can damage galvos
	c := color.RGBA{0x66, 0x33, 0x00, 0xFF}
	maxrad := 10260 * 2
	rad := maxrad
	frame := 0
	grow := false

	for {
		if rad <= 1 {
			grow = true
		} else if rad >= maxrad {
			grow = false
		}
		if grow {
			rad += 10
		} else {
			rad -= 10
		}
		for _, i := range xrange(0, pstep, 1) {
			f := float64(i) / float64(pstep) * 2.0 * math.Pi
			x := int(math.Cos(f) * float64(rad))
			y := int(math.Sin(f) * float64(rad))
			w.Write(etherdream.NewPoint(x, y, c).Encode())
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
