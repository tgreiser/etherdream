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
	"encoding/binary"
	"image/color"
	"io"
)

// PointStream is the interface clients should implement to
// generate points
type PointStream func(w io.WriteCloser)

// Point is a step in the laser stream, X, Y, RGB, Intensity and
// some other fields.
type Point struct {
	X     int16
	Y     int16
	R     uint16
	G     uint16
	B     uint16
	I     uint16
	U1    uint16
	U2    uint16
	Flags uint16
}

// NewPoint wil instantiate a point from the basic attributes.
func NewPoint(x, y int, c color.Color) *Point {
	r, g, b, a := c.RGBA()
	return &Point{
		X: int16(x),
		Y: int16(y),
		R: uint16(r),
		G: uint16(g),
		B: uint16(b),
		I: uint16(a),
	}
}

// Encode color values into a 18 byte struct Point
//
// Values must be specified for x, y, r, g, and b. If a value is not
// passed in for the other fields, i will default to max(r, g, b); the
// rest default to zero.
func (p Point) Encode() []byte {
	mut.Lock()
	if p.I <= 0 {
		p.I = p.R
		if p.G > p.I {
			p.I = p.G
		}
		if p.B > p.I {
			p.I = p.B
		}
	}
	var enc = make([]byte, 18)

	binary.LittleEndian.PutUint16(enc[0:2], p.Flags)
	// X and Y are actualy int16
	binary.LittleEndian.PutUint16(enc[2:4], uint16(p.X))
	binary.LittleEndian.PutUint16(enc[4:6], uint16(p.Y))

	binary.LittleEndian.PutUint16(enc[6:8], p.R)
	binary.LittleEndian.PutUint16(enc[8:10], p.G)
	binary.LittleEndian.PutUint16(enc[10:12], p.B)
	binary.LittleEndian.PutUint16(enc[12:14], p.I)
	binary.LittleEndian.PutUint16(enc[14:16], p.U1)
	binary.LittleEndian.PutUint16(enc[16:18], p.U2)
	mut.Unlock()
	return enc
}

// Points - Point list
type Points struct {
	Points []Point
}
