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

	dac.Play(pointStream)
}

func line(x, y, z, x2, y2, z2 float64) ln.Path {
	return ln.Path{ln.Vector{x, y, z}, ln.Vector{x2, y2, z2}}
}

func pointStream(w io.WriteCloser) {
	defer w.Close()

	c1 := color.RGBA{0x88, 0x00, 0x77, 0xFF}
	c2 := color.RGBA{0x00, 0x88, 0x00, 0xFF}
	frame := 0

	for {

		// compute 2D paths that depict the 3D scene
		paths := ln.Paths{
			line(0, 0, 0, 0, 5000, 0),
			line(5000, 0, 0, 5000, 5000, 0),
			line(10000, 0, 0, 10000, 5000, 0),
			line(12000, 0, 0, 12000, 5000, 0),
			line(14000, 0, 0, 14000, 5000, 0),
			line(14000, 5000, 0, 0, 5000, 0),
			line(0, 0, 0, 14000, 0, 0),
		}

		lp := len(paths)
		for iX := 0; iX < lp; iX++ {
			p := paths[iX]
			p2 := paths[0]
			if iX+1 < lp {
				p2 = paths[iX+1]
			}

			c := c1
			if iX%2 == 0 {
				c = c2
			}
			etherdream.DrawPath(w, p, c, 0.0)
			if p2[0].Distance(p[1]) > 0 {
				etherdream.BlankPath(w, ln.Path{p[1], p2[0]})
			}
		}

		frame++
	}
}
