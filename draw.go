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
var PreBlankCount = 0

// PostBlankCount is the number of blank samples to insert after moving
var PostBlankCount = 20

// DrawSpeed affects how many points will be sampled on your lines. Lower is
// more precise, but is more likely to flicker. Higher values will give smoother
// playback, but there may be gaps around corners. Try values 25-100.
var DrawSpeed = 50.0

// NumberOfSegments to use when interpolating the path
func NumberOfSegments(p ln.Path, drawSpeed float64) float64 {
	return p[0].Distance(p[1]) / drawSpeed
}

// DrawPath will use linear interpolation to draw fn+1 points along the path (fn segments)
// qual will override the LineQuality (see above).
func DrawPath(w io.WriteCloser, p ln.Path, c color.Color, drawSpeed float64) {
	if drawSpeed == 0.0 {
		drawSpeed = DrawSpeed
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

// BlankPath will necessary pauses to effectively blank a path
func BlankPath(w io.WriteCloser, p ln.Path) {
	for i := 1; i <= PreBlankCount; i++ {
		w.Write(NewPoint(int(p[0].X), int(p[0].Y), BlankColor).Encode())
	}

	for i := 1; i <= PostBlankCount; i++ {
		w.Write(NewPoint(int(p[1].X), int(p[1].Y), BlankColor).Encode())
	}
}
