package lib

import (
        "encoding/binary"
        "net"
)

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

func FindFirst() (*net.UDPAddr, *BroadcastPacket, error) {
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
