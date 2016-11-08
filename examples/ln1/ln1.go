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

func pointStream(w *io.PipeWriter) {
	defer w.Close()

	// create a scene and add a single cube
	scene := ln.Scene{}
	scene.Add(ln.NewCube(ln.Vector{-1, -1, -1}, ln.Vector{1, 1, 1}))

	// define camera parameters
	eye := ln.Vector{4, 3, 2}    // camera position
	center := ln.Vector{0, 0, 0} // camera looks at
	up := ln.Vector{0, 0, 1}     // up direction

	// define rendering parameters
	width := 10240.0  // rendered width
	height := 10240.0 // rendered height
	fovy := 50.0      // vertical field of view, degrees
	znear := 0.1      // near z plane
	zfar := 10.0      // far z plane
	step := 0.01      // how finely to chop the paths for visibility testing

	cmax := etherdream.ScaleColor(.5)
	frame := 0

	for {

		// compute 2D paths that depict the 3D scene
		paths := scene.Render(eye, center, up, width, height, fovy, znear, zfar, step)

		lp := len(paths)
		for iX := 0; iX < lp; iX++ {
			p := paths[iX]
			p2 := paths[0]
			if iX+1 < lp {
				p2 = paths[iX+1]
			}
			//log.Printf("%v - %v\n", p, cmax)
			etherdream.DrawPath(w, p, cmax)
			etherdream.BlankPath(w, ln.Path{p[1], p2[0]})
			//w.Write(etherdream.NewPoint(pt.X, pt.Y, cmax, cmax, cmax, cmax).Encode())
		}

		frame++
	}
}
