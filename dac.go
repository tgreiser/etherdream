package etherdream

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"
)

var mut = &sync.Mutex{}

// PointSize is the number of bytes in a point struct
const PointSize uint16 = 18

type ProtocolError struct {
	Msg string
}

func (e *ProtocolError) Error() string {
	return e.Msg
}

type DAC struct {
	Host           string
	Port           string
	FirmwareString string
	LastStatus     *DACStatus
	Reader         *io.PipeReader
	Writer         *io.PipeWriter
	buf            bytes.Buffer
	conn           net.Conn
	r              io.Reader
}

func NewDAC(host string) DAC {
	// connect to the DAC over TCP
	r, w := io.Pipe()
	return DAC{Host: host, Port: "7765", Reader: r, Writer: w}
}

func (d *DAC) Close() {
	d.conn.Close()
}

func (d *DAC) Init() error {
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
	fmt.Printf("Firmware String: %v\n", d.FirmwareString)
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

func (d *DAC) ReadResponse(cmd string) (*DACStatus, error) {
	data, err := d.Read(22)
	if err != nil {
		fmt.Errorf("%v\n", err)
		return nil, err
	}

	resp := data[0]
	cmdR := data[1]
	status := NewDACStatus(data[2:])
	fmt.Printf("Read response: %s %s\n", string(resp), string(cmdR))

	if cmdR != []byte(cmd)[0] {
		return nil, &ProtocolError{fmt.Sprintf("Expected resp for %s, got %s", string(cmd), string(cmdR))}
	}
	if resp != []byte("a")[0] {
		return nil, &ProtocolError{fmt.Sprintf("Expected ACK, got %s", string(resp))}
	}
	d.LastStatus = status
	return status, nil
}

func (d DAC) Send(cmd []byte) error {
	_, err := d.conn.Write(cmd)
	return err
}

func (d DAC) Begin(lwm uint16, rate uint32) (*DACStatus, error) {
	var cmd []byte = make([]byte, 7)
	cmd[0] = []byte("b")[0]
	binary.LittleEndian.PutUint16(cmd[1:3], lwm)
	binary.LittleEndian.PutUint32(cmd[3:7], rate)
	d.Send(cmd)
	s, err := d.ReadResponse("b")
	fmt.Println(s)
	return s, err
}

func (d DAC) Update(lwm uint16, rate uint32) (*DACStatus, error) {
	var cmd []byte = make([]byte, 7)
	cmd[0] = []byte("u")[0]
	binary.LittleEndian.PutUint16(cmd[1:3], lwm)
	binary.LittleEndian.PutUint32(cmd[3:7], rate)
	d.Send(cmd)
	return d.ReadResponse("u")
}

func (d DAC) Write(b []byte) (*DACStatus, error) {
	l := uint16(len(b))
	cmd := make([]byte, l+3)
	cmd[0] = []byte("d")[0]
	binary.LittleEndian.PutUint16(cmd[1:3], l/PointSize)
	copy(cmd[3:], b)

	fmt.Printf("Writing cmd - length: 3 + %v\n", l)
	d.Send(cmd)
	fmt.Println("Reading response d")
	return d.ReadResponse("d")
}

// Prepare command
func (d DAC) Prepare() (*DACStatus, error) {
	d.Send([]byte("p"))
	return d.ReadResponse("p")
}

// Stop command
func (d DAC) Stop() (*DACStatus, error) {
	d.Send([]byte("s"))
	return d.ReadResponse("s")
}

// Emergency Stop command
func (d DAC) EmergencyStop() (*DACStatus, error) {
	d.Send([]byte("\xFF"))
	return d.ReadResponse("\xFF")
}

// Clear Emergency Stop command
func (d DAC) ClearEmergencyStop() (*DACStatus, error) {
	d.Send([]byte("c"))
	return d.ReadResponse("c")
}

// Ping command
func (d DAC) Ping() (*DACStatus, error) {
	d.Send([]byte("?"))
	return d.ReadResponse("?")
}

// Start playing a stream generator and sending output to the laser
func (d DAC) Play(stream PointStream) {
	// First, prepare the stream
	if d.LastStatus.PlaybackState == 2 {
		fmt.Errorf("Already playing?!")
	} else if d.LastStatus.PlaybackState == 0 {
		d.Prepare()
		fmt.Printf("DAC prepared: %v\n", d.LastStatus)
	}

	started := 0
	// Start stream
	go stream(d.Writer)

	for {
		// Read calls from the pipe
		cap := 1799 - d.LastStatus.BufferFullness

		if cap < 100 {
			time.Sleep(time.Millisecond * 5)
			continue
		}

		by := make([]byte, cap*PointSize)
		idx := 0
		payloadSize := int(cap)

		fmt.Printf("Buffer capacity: %v pts\n", cap)

		for idx < payloadSize {
			_, err := d.Reader.Read(by[idx:])
			if err != nil {
				fmt.Printf("Error playing stream: %v", err)
				continue
			}
			idx++
			//fmt.Printf("Read %v bytes from pipe. Cap: %v / %v\n", ln, idx, cap)

		}

		mut.Lock()
		t0 := time.Now()
		d.Write(by)
		t1 := time.Now()
		fmt.Printf("%v bytes took %v\n", len(by), t1.Sub(t0).String())

		if started == 0 {
			d.Begin(0, 30000)
			started = 1
			fmt.Println("Begin executed")
		}
		fmt.Printf("Status: %v\n", d.LastStatus)
		mut.Unlock()
		runtime.Gosched()

	}
}

type PointStream func(w *io.PipeWriter) Points

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

func NewPoint(x, y, r, g, b, i int) *Point {
	return &Point{
		X: int16(x),
		Y: int16(y),
		R: uint16(r),
		G: uint16(g),
		B: uint16(b),
		I: uint16(i),
	}
}

// Pack color values into a 18 byte struct Point
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
	var enc []byte = make([]byte, 18)

	//fmt.Printf("Encoding %v\n", p)

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

type Points struct {
	Points []Point
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
