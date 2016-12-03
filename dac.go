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
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var mut = &sync.Mutex{}

// PointSize is the number of bytes in a point struct
const PointSize uint16 = 18

// ProtocolError indicates a protocol level error. I've
// never seen one, but maybe you will.
type ProtocolError struct {
	Msg string
}

func (e *ProtocolError) Error() string {
	return e.Msg
}

// DAC is the interface to the Ether Dream Digital to
// Analog Converter that turns network signals into
// ILDA control singnals for a laser scanner.
type DAC struct {
	Host           string
	Port           string
	FirmwareString string
	LastStatus     *DACStatus
	Reader         *io.PipeReader
	Writer         *io.PipeWriter
	PointsPlayed   int
	buf            bytes.Buffer
	conn           net.Conn
	r              io.Reader
}

// NewDAC will connect to an Ether Dream device over TCP
// or it will return an error
func NewDAC(host string) (*DAC, error) {
	// connect to the DAC over TCP
	r, w := io.Pipe()
	dac := &DAC{Host: host, Port: "7765", Reader: r, Writer: w}
	err := dac.init()
	return dac, err
}

// Close the network connection, you should. -- Yoda
func (d *DAC) Close() {
	d.conn.Close()
}

func (d *DAC) init() error {
	c, err := net.DialTimeout("tcp", d.Host+":"+d.Port, time.Second*15)
	if err != nil {
		return err
	}
	d.conn = c

	_, err = d.ReadResponse("?")
	if err != nil {
		return err
	}

	d.Send([]byte("v"))
	by, err2 := d.Read(32)
	if err2 != nil {
		return err2
	}

	d.FirmwareString = strings.TrimSpace(strings.Replace(string(by), "\x00", " ", -1))

	return nil
}

func (d *DAC) Read(l int) ([]byte, error) {
	if l > d.buf.Len() {
		// read more bytes into the buffer
		_, err := io.CopyN(&d.buf, d.conn, int64(l))
		if err != nil {
			return []byte{}, err
		}
	}
	ret := make([]byte, l)
	_, err := d.buf.Read(ret)
	return ret, err
}

// ReadResponse reads the ACK/NACK response to a command
func (d *DAC) ReadResponse(cmd string) (*DACStatus, error) {
	data, err := d.Read(22)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil, err
	}

	resp := data[0]
	cmdR := data[1]
	status := NewDACStatus(data[2:])
	//fmt.Printf("\nRead response: %s %s\n", string(resp), string(cmdR))

	if cmdR != []byte(cmd)[0] {
		return nil, &ProtocolError{fmt.Sprintf("Expected resp for %s, got %s", string(cmd), string(cmdR))}
	}
	if resp != []byte("a")[0] {
		return nil, &ProtocolError{fmt.Sprintf("Expected ACK, got %s", string(resp))}
	}
	d.LastStatus = status
	return status, nil
}

// Send a command to the DAC
func (d DAC) Send(cmd []byte) error {
	_, err := d.conn.Write(cmd)
	return err
}

// Begin Playback
// This causes the DAC to begin producing output. lwm is
// currently unused. rate is the number of points per second
// to be read from the buffer. If the playback system was
// Prepared and there was data in the buffer, then the DAC
// will reply with ACK; otherwise, it replies with NAK - Invalid.
func (d *DAC) Begin(lwm uint16, rate uint32) (*DACStatus, error) {
	var cmd = make([]byte, 7)
	cmd[0] = []byte("b")[0]
	binary.LittleEndian.PutUint16(cmd[1:3], lwm)
	binary.LittleEndian.PutUint32(cmd[3:7], rate)
	d.Send(cmd)
	s, err := d.ReadResponse("b")
	fmt.Printf("Begin: %v\n\n", s)
	return s, err
}

// Update should not exist?
// Maybe this is the 'q' command now.
func (d *DAC) Update(lwm uint16, rate uint32) (*DACStatus, error) {
	var cmd = make([]byte, 7)
	cmd[0] = []byte("u")[0]
	binary.LittleEndian.PutUint16(cmd[1:3], lwm)
	binary.LittleEndian.PutUint32(cmd[3:7], rate)
	d.Send(cmd)
	return d.ReadResponse("u")
}

func (d *DAC) Write(b []byte) (*DACStatus, error) {
	l := uint16(len(b))
	cmd := make([]byte, l+3)
	cmd[0] = []byte("d")[0]
	binary.LittleEndian.PutUint16(cmd[1:3], l/PointSize)
	copy(cmd[3:], b)

	d.Send(cmd)
	return d.ReadResponse("d")
}

// Prepare command
func (d *DAC) Prepare() (*DACStatus, error) {
	d.Send([]byte("p"))
	return d.ReadResponse("p")
}

// Stop command
func (d *DAC) Stop() (*DACStatus, error) {
	d.Send([]byte("s"))
	return d.ReadResponse("s")
}

// EmergencyStop command causes the light engine to
// enter the E-Stop state, regardless of its previous
// state. It is always ACKed.
func (d *DAC) EmergencyStop() (*DACStatus, error) {
	d.Send([]byte("\xFF"))
	return d.ReadResponse("\xFF")
}

// ClearEmergencyStop command. If the light engine was in
// E-Stop state due to an emergency stop command (either from
// a local stop condition or over the network), then this
// command resets it to be Ready. It is ACKed if the DAC was
// previously in E-Stop; otherwise it is replied to with a NAK -
// Invalid. If the condition that caused the emergency stop is
// still active (E-Stop input still asserted, temperature still
// out of bounds, etc.), then a NAK - Stop Condition is sent.
func (d *DAC) ClearEmergencyStop() (*DACStatus, error) {
	d.Send([]byte("c"))
	return d.ReadResponse("c")
}

// Ping command
func (d *DAC) Ping() (*DACStatus, error) {
	d.Send([]byte("?"))
	return d.ReadResponse("?")
}

// ShouldPrepare or not? State 1 and 2 are good. Some Flags
// need prepare to reset an invalid state.
func (d DAC) ShouldPrepare() bool {
	return d.LastStatus.PlaybackState == 0 ||
		d.LastStatus.PlaybackFlags&2 == 2 ||
		d.LastStatus.PlaybackFlags&4 == 4
}

// Measure how long it takes to play 10,000 points
func (d *DAC) Measure(stream PointStream) {
	t0 := time.Now()

	go d.Play(stream, true)

	for {
		if d.PointsPlayed >= 100000 {
			t1 := time.Now()
			fmt.Printf("%v took %v\n", d.PointsPlayed, t1.Sub(t0).String())
			os.Exit(0)
		}
		runtime.Gosched()
	}
}

// Play a stream generator and begin sending output to the laser
func (d *DAC) Play(stream PointStream, debug bool) {
	// First, prepare the stream
	if d.LastStatus.PlaybackState == 2 {
		if debug {
			fmt.Printf("Error: Already playing?!")
		}
	} else if d.ShouldPrepare() {
		st, err := d.Prepare()
		if err != nil {
			fmt.Printf("ERROR: Failed to prepare: %v\n\n", err)
		}
		if debug {
			fmt.Printf("DAC prepared: %v\n\n", st)
		}
	}

	started := 0
	// Start stream
	go stream(d.Writer)

	for {
		// Read calls from the pipe
		cap := 1799 - d.LastStatus.BufferFullness

		if cap < 100 {
			time.Sleep(time.Millisecond * 5)
			d.Ping()
			continue
		}

		by := make([]byte, cap*PointSize)
		idx := 0
		payloadSize := int(cap)

		if debug {
			fmt.Printf("Buffer capacity: %v pts\n", cap)
		}

		for idx < payloadSize {
			bdx := idx * int(PointSize)
			_, err := d.Reader.Read(by[bdx:])
			if err != nil {
				fmt.Printf("Error playing stream: %v", err)
				break
			}
			idx++

		}

		mut.Lock()
		st, err := d.Write(by)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
		}

		d.PointsPlayed += len(by) / int(PointSize)
		if debug {
			fmt.Printf("Points: %v\nStatus: %v\n", d.PointsPlayed, st)
		}

		if started == 0 {
			st, err := d.Begin(0, 30000)
			if err != nil {
				fmt.Printf("ERROR on Begin: %v\n\n", err)
			}
			started = 1
			if debug {
				fmt.Printf("\nBegin executed: %v\n", st)
			}
		}
		if debug {
			fmt.Println()
		}

		mut.Unlock()
		runtime.Gosched()

	}
}

// FindFirstDAC starts a UDP server to listen for broadcast packets on your network. Return the UDPAddr
// of the first Ether Dream DAC located
func FindFirstDAC() (*net.UDPAddr, *BroadcastPacket, error) {
	// listen for broadcast packets
	sock, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 7654,
	})

	data := make([]byte, 1024)
	_, addr, err := sock.ReadFromUDP(data)
	if err != nil {
		return nil, nil, err
	}

	bp := NewBroadcastPacket(data)
	return addr, bp, nil
}
