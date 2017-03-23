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

	dac.Play(pointStream)
}

func pointStream(w io.WriteCloser) {
	defer w.Close()

	// Don't use a low # of steps, 30 and below can damage galvos
	// We'll use the number of points in a frame for optimal sampling
	pstep := etherdream.FramePoints()
	c := color.RGBA{0x66, 0x33, 0x00, 0xFF}
	maxrad := 10260
	rad := maxrad
	grow := false

	for {
		if rad <= 1 {
			grow = true
		} else if rad >= maxrad {
			grow = false
		}
		if grow {
			rad += 100
		} else {
			rad -= 100
		}
		var pt *etherdream.Point
		for i := 0; i < pstep; i++ {
			f := float64(i) / float64(pstep) * 2.0 * math.Pi
			x := int(math.Cos(f) * float64(rad))
			y := int(math.Sin(f) * float64(rad))
			pt = etherdream.NewPoint(x, y, c)
			w.Write(pt.Encode())
		}

		_ = etherdream.NextFrame(w, pstep, *pt)
	}
}
