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
var ratio = flag.Float64("ratio", 3.0, "Ratio of inner to outer")
var amax = flag.Float64("max-diam", 12000.0, "Maximum diameter of the outer circle (a)")
var amin = flag.Float64("min-diam", 500.0, "Minimum diameter of the outer circle (b)")
var h = flag.Float64("point-dist", 4000.0, "Distance from the draw point to the center of the inner circle (h)")

func pointStream(w io.WriteCloser) {
	defer w.Close()
	pcount := 380
	//etherdream.FramePoints() - *etherdream.BlankCount
	log.Printf("PCount %v\n", pcount)
	grow := true

	//var pt *etherdream.Point
	a := *amin
	b := a / *ratio
	//pt := etherdream.NewPoint(0,0, color.Black)
	for {
		for iY := 1; iY < 3; iY++ {
			for iX := 0; iX < pcount; iX++ {
				graph(w, pcount, iX, iY, a/float64(iY), b/float64(iY), *h/float64(iY))
			}
		}
		//etherdream.NextFrame(w, pcount, *pt)
		if grow {
			a = a / .9
		} else {
			a = a * .9
		}
		b = a / *ratio
		if a < *amin {
			grow = true
		}
		if a > *amax {
			grow = false
		}
	}
}

var xyMin = -5600
var xyMax = 5600
var xyRange = xyMax - xyMin

func graph(w io.WriteCloser, max, cur, rep int, a, b, h float64) *etherdream.Point {
	th := float64(cur) / 10.0
	x := (a-b)*math.Cos(th) + h*math.Cos((a-b)/b*th)
	y := (a-b)*math.Sin(th) - h*math.Sin((a-b)/b*th)

	pt := etherdream.NewPoint(int(x), int(y), colorOsc(cur, rep))
	w.Write(pt.Encode())
	return pt
}

func colorOsc(cur, rep int) color.Color {
	// blank the beginning of the frame also
	if cur < *etherdream.BlankCount {
		return etherdream.BlankColor
	}
	if rep == 2 {
		return color.RGBA{0x00, 0x55, 0x55, 0x77}
	}
	return color.RGBA{0x33, 0x00, 0x55, 0xff}
}
