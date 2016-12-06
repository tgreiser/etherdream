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
	"math/rand"

	"image/color"

	"flag"

	"github.com/tgreiser/etherdream"
	"github.com/tgreiser/ln/ln"
)

var speed = flag.Float64("draw-speed", 50.0, "Draw speed (25-100). Lower is more precision but slower.")

func main() {
	flag.Parse()
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

func cube(x, y, z float64) ln.Shape {
	size := 0.5
	v := ln.Vector{x, y, z}
	return ln.NewCube(v.SubScalar(size), v.AddScalar(size))
}

func pointStream(w io.WriteCloser) {
	defer w.Close()

	scene := ln.Scene{}
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			z := rand.Float64()
			scene.Add(cube(float64(x), float64(y), z))
		}
	}
	eye := ln.Vector{6, 5, 3}
	center := ln.Vector{0, 0, 0}
	up := ln.Vector{0, 0, 1}
	width := 10000.0
	height := 10000.0
	fovy := 30.0

	c := color.RGBA{0x77, 0x00, 0x00, 0xFF}
	frame := 0

	for {

		// compute 2D paths that depict the 3D scene
		paths := scene.Render(eye, center, up, width, height, fovy, 0.1, 100, 0.01)
		paths.Optimize()

		lp := len(paths)
		for iX := 0; iX < lp; iX++ {
			p := paths[iX]
			p2 := paths[0]
			if iX+1 < lp {
				p2 = paths[iX+1]
			}
			etherdream.DrawPath(w, p, c, *speed)
			if p2[0].Distance(p[1]) > 0 {
				etherdream.BlankPath(w, ln.Path{p[1], p2[0]})
			}
		}

		frame++
	}
}
