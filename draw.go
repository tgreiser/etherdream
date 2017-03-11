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

package etherdream

import (
	"flag"
	"fmt"
	"image/color"
	"io"
	"log"
	"time"

	"github.com/tgreiser/ln/ln"
)

// BlankCount is the number of blank samples to insert after moving
var BlankCount = flag.Int("blank-count", 20, "How many samples to wait after drawing a blanking line.")

// DrawSpeed affects how many points will be sampled on your lines. Lower is
// more precise, but is more likely to flicker. Higher values will give smoother
// playback, but there may be gaps around corners. Try values 25-100.
var DrawSpeed = flag.Float64("draw-speed", 50.0, "Draw speed (25-100). Lower is more precision but slower.")

// Debug mode
var Debug = flag.Bool("debug", false, "Enable debug output.")

// Dump will output the point stream coordinates
var Dump = flag.Bool("dump", false, "Dump point stream to stdout.")

var tf0 = time.Now()

// NextFrame advances playback ... add some blank points
func NextFrame(w io.WriteCloser, pointsPlayed int, last Point) int {
	times := framePoints - pointsPlayed
	by := NewPoint(int(last.X), int(last.Y), BlankColor).Encode()
	for iX := 0; iX < times; iX++ {
		w.Write(by)
	}
	if *Debug {
		tf1 := time.Now()
		log.Printf("%v - Frame %v added %v empty points", tf1.Sub(tf0), frameCount, times)
		tf0 = tf1
	}
	if *Dump {
		fmt.Printf("---- %v x %v\t%v\t%v\t%v\t%v\n", times, last.X, last.Y, 0, 0, 0)
	}
	frameCount++
	return frameCount
}

// NumberOfSegments to use when interpolating the path
func NumberOfSegments(p ln.Path, drawSpeed float64) float64 {
	return p[0].Distance(p[1]) / drawSpeed
}

// DrawPath will use linear interpolation to draw fn+1 points along the path (fn segments)
// qual will override the LineQuality (see above).
func DrawPath(w io.WriteCloser, p ln.Path, c color.Color, drawSpeed float64) {
	if drawSpeed == 0.0 {
		drawSpeed = *DrawSpeed
	}
	dist := p[1].Sub(p[0])

	fn := NumberOfSegments(p, drawSpeed)
	for iX := 0.0; iX < fn; iX++ {
		x := dist.X * (iX / fn)
		y := dist.Y * (iX / fn)
		np := p[0].Add(ln.Vector{x, y, 0})
		w.Write(NewPoint(int(np.X), int(np.Y), c).Encode())
	}
	w.Write(NewPoint(int(p[1].X), int(p[1].Y), c).Encode())
}

// BlankPath will add the necessary pause to effectively blank a path
func BlankPath(w io.WriteCloser, p ln.Path) {
	for i := 1; i <= *BlankCount; i++ {
		w.Write(NewPoint(int(p[1].X), int(p[1].Y), BlankColor).Encode())
	}
}
