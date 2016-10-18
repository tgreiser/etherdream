package etherdream

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

type ProtocolError struct {
	Msg string
}

func (e *ProtocolError) Error() string {
	return e.Msg
}

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

type DAC struct {
	Host           string
	Port           string
	FirmwareString string
	LastStatus     *DACStatus
	buf            bytes.Buffer
	conn           net.Conn
	r              io.Reader
}

func NewDAC(host string) DAC {
	// connect to the DAC over TCP
	return DAC{Host: host, Port: "7765"}
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

	status, err := d.ReadResponse("?")
	if err != nil {
		return err
	}
	fmt.Println(status)

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

func (d DAC) ReadResponse(cmd string) (*DACStatus, error) {
	data, err := d.Read(22)
	if err != nil {
		return nil, err
	}

	resp := data[0]
	cmdR := data[1]
	status := NewDACStatus(data[2:])

	if cmdR != []byte(cmd)[0] {
		return nil, &ProtocolError{fmt.Sprintf("Expected resp for %r, got %r", cmd, cmdR)}
	}
	if resp != []byte("a")[0] {
		return nil, &ProtocolError{fmt.Sprintf("Expected ACK, got %r", resp)}
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
	return d.ReadResponse("b")
}

func (d DAC) Update(lwm uint16, rate uint32) (*DACStatus, error) {
	var cmd []byte = make([]byte, 7)
	cmd[0] = []byte("u")[0]
	binary.LittleEndian.PutUint16(cmd[1:3], lwm)
	binary.LittleEndian.PutUint32(cmd[3:7], rate)
	d.Send(cmd)
	return d.ReadResponse("u")
}

// Pack color values into a struct Point
//
// Values must be specified for x, y, r, g, and b. If a value is not
// passed in for the other fields, i will default to max(r, g, b); the
// rest default to zero.
func (d DAC) EncodePoint(p Point) []byte {
	if p.I <= 0 {
		p.I = p.R
		if p.G > p.I {
			p.I = p.G
		}
		if p.B > p.I {
			p.I = p.B
		}
	}
	var enc []byte = make([]byte, 16)

	binary.LittleEndian.PutUint16(enc[0:2], p.Flags)
	enc[2] = p.X
	enc[3] = p.Y
	binary.LittleEndian.PutUint16(enc[4:6], p.R)
	binary.LittleEndian.PutUint16(enc[6:8], p.G)
	binary.LittleEndian.PutUint16(enc[8:10], p.B)
	binary.LittleEndian.PutUint16(enc[10:12], p.I)
	binary.LittleEndian.PutUint16(enc[12:14], p.U1)
	binary.LittleEndian.PutUint16(enc[14:16], p.U2)
	return enc
}

func (d DAC) Write(pts Points) (*DACStatus, error) {
	l := uint16(16 * len(pts.Points))
	cmd := make([]byte, l+3)
	cmd[0] = []byte("d")[0]
	binary.LittleEndian.PutUint16(cmd[1:3], l)
	for i, p := range pts.Points {
		copy(cmd[i+3:i+3+16], d.EncodePoint(p))
	}

	d.Send(cmd)
	return d.ReadResponse("d")
}

func (d DAC) Prepare() (*DACStatus, error) {
	d.Send([]byte("p"))
	return d.ReadResponse("p")
}

func (d DAC) Stop() (*DACStatus, error) {
	d.Send([]byte("s"))
	return d.ReadResponse("s")
}

func (d DAC) EmergencyStop() (*DACStatus, error) {
	d.Send([]byte("\xFF"))
	return d.ReadResponse("\xFF")
}

func (d DAC) ClearEmergencyStop() (*DACStatus, error) {
	d.Send([]byte("c"))
	return d.ReadResponse("c")
}

func (d DAC) Ping() (*DACStatus, error) {
	d.Send([]byte("?"))
	return d.ReadResponse("?")
}

/*
func (d DAC) PlayStream(stream) {
}
*/

type Point struct {
	X     uint8
	Y     uint8
	R     uint16
	G     uint16
	B     uint16
	I     uint16
	U1    uint16
	U2    uint16
	Flags uint16
}

type Points struct {
	Points []Point
}

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
