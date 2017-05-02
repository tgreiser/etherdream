/*
# Copyright 2016 Tim Greiser

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
	"image/color"
	"io"
	"log"
	"math"

	"flag"

	"github.com/tgreiser/etherdream"
	"github.com/tgreiser/ln/ln"
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

	dac.Play(pointStream)
}

var max = flag.Int("speed", 500, "Speed to run the oscillation (1-20000)")
var xAmp = flag.Float64("x-amp", 20.0, "X amplitude (1.0 - 50.0)")
var yAmp = flag.Float64("y-amp", 5.0, "Y amplitude (1.0 - 50.0)")

func pointStream(w io.WriteCloser) {
	defer w.Close()
	pcount := etherdream.FramePoints() - *etherdream.BlankCount
	xx := 0
	xy := 0

	var pt *etherdream.Point
	for {
		xRate := etherdream.Osc(xx%*max, *max, *xAmp, 1, 2.5)
		yRate := etherdream.Osc(xy%*max, *max, *yAmp, 1, 2.5)
		xx++
		xy++
		for iX := 0; iX < pcount; iX++ {
			if iX == 0 && pt != nil {
				nextPt := graph(w, pcount, iX, xRate, yRate)
				blank := ln.Path{pt.ToVector(), nextPt.ToVector()}
				etherdream.BlankPath(w, blank)
				pt = nextPt
			} else {
				pt = graph(w, pcount, iX, xRate, yRate)
			}
		}
	}
}

var xyMin = -5600
var xyMax = 5600
var xyRange = xyMax - xyMin

func graph(w io.WriteCloser, max, cur int, xRate, yRate float64) *etherdream.Point {
	// 0 - max = 0 - 2 radians
	rad := float64(cur+1) / float64(max) * 2.0 * math.Pi
	x := math.Sin(rad*xRate)*float64(xyRange) + float64(xyMin)
	y := math.Sin(rad*yRate)*float64(xyRange) + float64(xyMin)

	pt := etherdream.NewPoint(int(x), int(y), colorOsc(max, cur))
	w.Write(pt.Encode())
	return pt
}

func colorOsc(max, cur int) color.Color {
	return color.RGBA{0x00, 0x00, 0x55, 0xff}
}
