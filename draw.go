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
	"image/color"
	"io"

	"github.com/tgreiser/ln/ln"
)

// PreBlankCount is the number of blank samples to insert before moving
var PreBlankCount = 1

// PostBlankCount is the number of blank samples to insert after moving
var PostBlankCount = 25

// LineStepSize is the length of a segment that can be drawn in one time unit - used to chop the lines
var LineStepSize = 35.0

// LineSteps is the distance of each step when moving from p[0] to p[1]
func LineSteps(p ln.Path) float64 {
	return LineStepSize
}

// LineQuality controls the resolution of your lines. Lower is more precise, but more flicker. Higher
// will be smoother playback, but there may be gaps around corners. Try values 30-80.
var LineQuality = 100.0

// LerpSegments returns the number of segments to interpolate the path over
func LerpSegments(p ln.Path, quality float64) float64 {
	return p[0].Distance(p[1]) / quality
}

// DrawPath will use linear interpolation to draw fn+1 points along the path (fn segments)
// qual will override the LineQuality (see above).
func DrawPath(w *io.PipeWriter, p ln.Path, c color.Color, qual float64) {
	if qual == 0.0 {
		qual = LineQuality
	}
	dist := p[1].Sub(p[0])

	fn := LerpSegments(p, qual)
	for iX := 0.0; iX < fn; iX++ {
		x := dist.X * (iX / fn)
		y := dist.Y * (iX / fn)
		np := p[0].Add(ln.Vector{x, y, 0})
		w.Write(NewPoint(int(np.X), int(np.Y), c).Encode())
	}
	w.Write(NewPoint(int(p[1].X), int(p[1].Y), c).Encode())
}

// ChopPath is the deprecated way to slice up the paths
func ChopPath(w *io.PipeWriter, p ln.Path, c color.Color) {
	step := LineSteps(p)
	for _, pt := range p.Chop(step) {
		w.Write(NewPoint(int(pt.X), int(pt.Y), c).Encode())
	}
}

// BlankPath will blank points on a path
func BlankPath(w *io.PipeWriter, p ln.Path) {
	for i := 1; i <= PreBlankCount; i++ {
		w.Write(NewPoint(int(p[0].X), int(p[0].Y), BlankColor).Encode())
	}

	/*
		for _, pt := range p.Chop(LineSteps(p)) {
			w.Write(NewPoint(int(pt.X), int(pt.Y), BlankColor).Encode())
		}
	*/
	for i := 1; i <= PostBlankCount; i++ {
		w.Write(NewPoint(int(p[1].X), int(p[1].Y), BlankColor).Encode())
	}
}
