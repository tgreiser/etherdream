/*
# Copyright 2016 Tim Greiser
# Based on work by Jacob Potter, some comments are from his
# protocol documents

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
	"fmt"
)

// DACStatus is a struct of status informaion sent by the etherdream DAC
type DACStatus struct {
	Protocol         uint8
	LightEngineState uint8
	PlaybackState    uint8
	Source           uint8
	LightEngineFlags uint16
	PlaybackFlags    uint16
	SourceFlags      uint16
	BufferFullness   uint16
	PointRate        uint32
	PointCount       uint32
}

func NewDACStatus(b []byte) *DACStatus {
	return &DACStatus{
		Protocol:         b[0],
		LightEngineState: b[1],
		PlaybackState:    b[2],
		Source:           b[3],
		LightEngineFlags: binary.LittleEndian.Uint16(b[4:6]),
		PlaybackFlags:    binary.LittleEndian.Uint16(b[6:8]),
		SourceFlags:      binary.LittleEndian.Uint16(b[8:10]),
		BufferFullness:   binary.LittleEndian.Uint16(b[10:12]),
		PointRate:        binary.LittleEndian.Uint32(b[12:16]),
		PointCount:       binary.LittleEndian.Uint32(b[16:20]),
	}
}
func (st DACStatus) String() string {
	return fmt.Sprintf("Light engine: state %d, flags 0x%x\n", st.LightEngineState, st.LightEngineFlags) +
		fmt.Sprintf("Playback: state %d, flags 0x%x\n", st.PlaybackState, st.PlaybackFlags) +
		fmt.Sprintf("Buffer: %d points\n", st.BufferFullness) +
		fmt.Sprintf("Playback: %d kpps, %d points played", st.PointRate, st.PointCount) +
		fmt.Sprintf("Source: %d, flags 0x%x", st.Source, st.SourceFlags)
}

// BroadcastPacket is the various capabilities advertised by the DAC
type BroadcastPacket struct {
	MAC            []uint8
	HWRev          uint16
	SWRev          uint16
	BufferCapacity uint16
	MaxPointRate   uint32
	Status         *DACStatus
}

func NewBroadcastPacket(b []byte) *BroadcastPacket {
	return &BroadcastPacket{
		MAC:            b[0:6],
		HWRev:          binary.LittleEndian.Uint16(b[6:8]),
		SWRev:          binary.LittleEndian.Uint16(b[8:10]),
		BufferCapacity: binary.LittleEndian.Uint16(b[10:12]),
		MaxPointRate:   binary.LittleEndian.Uint32(b[12:16]),
		Status:         NewDACStatus(b[16:36]),
	}
}

func (bp BroadcastPacket) String() string {
	return fmt.Sprintf("MAC: %02x %02x %02x %02x %02x %02x\n", bp.MAC[0], bp.MAC[1], bp.MAC[2], bp.MAC[3], bp.MAC[4], bp.MAC[5]) +
		fmt.Sprintf("HW %d, SW %d\n", bp.HWRev, bp.SWRev) +
		fmt.Sprintf("Capabilities: max %d points, %d kpps", bp.BufferCapacity, bp.MaxPointRate)
}
