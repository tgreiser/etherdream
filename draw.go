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
	"io"

	"github.com/tgreiser/ln/ln"
)

// DrawPath will interpolate and draw points along the path
func DrawPath(w *io.PipeWriter, p ln.Path, cmax int) {
	for _, pt := range p.Chop(30) {
		w.Write(NewPoint(int(pt.X), int(pt.Y), cmax, cmax, cmax, cmax).Encode())
	}
}

// BlankPath will blank points on a path
func BlankPath(w *io.PipeWriter, p ln.Path) {
	for i := 1; i <= 5; i++ {
		w.Write(NewPoint(int(p[0].X), int(p[0].Y), 0, 0, 0, 0).Encode())
	}
	for _, pt := range p.Chop(30) {
		w.Write(NewPoint(int(pt.X), int(pt.Y), 0, 0, 0, 0).Encode())
	}
	for i := 1; i <= 25; i++ {
		w.Write(NewPoint(int(p[1].X), int(p[1].Y), 0, 0, 0, 0).Encode())
	}
}

/*
// Blank points from xy1 to xy2
func Blank(w *io.PipeWriter, x1, y1, x2, y2 int) {
    Blank(w, float64(x1), float64(y1), float64(x2), float64(y2))
}
*/
