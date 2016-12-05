# Go Ether Dream

Go language interface to the Ether Dream laser DAC. Current features: blanking, basic path optimization, quality (trade resolution for frame rate) and 3D scene rendering via [ln](https://github.com/tgreiser/ln). For an introduction to laser projectors and programming, see: [Laser Hack 101 presentation slides](https://github.com/tgreiser/etherdream-touch-designer/blob/master/laser_hack_101.pdf).

Based on the work of [j4cbo](https://github.com/j4cbo/j4cDAC/), [echelon](https://github.com/echelon) and [fogleman](https://github.com/fogleman)

![Spiral](http://prim8.net/art/spiral.jpg)

## Setup

This assumes you are plugged in to your ether dream via ethernet cable. You
may need to set up some rules for your firewall. Inbound communications 
are needed for the initial broadcast signal and handshake, if you don't
need to Find a DAC, you can use outbound only.

- Outbound rule: TCP port 7765
- Inbound rule: UDP port 7654

## Install

Assuming you have Go set up and installed, just:

    go get github.com/tgreiser/etherdream
    # cd to the etherdream directory
    
You can run any of the examples like:

    go run examples/square/square.go
    
## Connecting

If you have opened the necessary ports, the Ether Dream will broadcast it's identity on the network. You just have to connect to it and provide a [PointStream](https://godoc.org/github.com/tgreiser/etherdream#PointStream)

    func main() {
        log.Printf("Listening...\n")
        addr, bp, err := etherdream.FindFirstDAC()
        if err != nil {
            log.Fatalf("Network error: %v", err)
        }

        log.Printf("Found DAC at %v\n", addr)
        log.Printf("BP: %v\n\n", bp)

        dac, err := etherdream.NewDAC(addr.IP.String())
        if err != nil {
            log.Fatal(err)
        }
        defer dac.Close()
        log.Printf("Initialized:  %v\n\n", dac.LastStatus)
        log.Printf("Firmware String: %v\n\n", dac.FirmwareString)

        debug := false
        dac.Play(squarePointStream, debug)
    }

## Point Streams

Point streams should contain an infinite loop that will use the PipeWriter interface to output the points to the DAC.

    func squarePointStream(w *io.PipeWriter) {
        defer w.Close()
        pmax := 15600
        pstep := 100
        for {
            for _, x := range xrange(-pmax, pmax, pstep) {
                w.Write(etherdream.NewPoint(x, pmax, color.RGBA{0xff, 0x00, 0x00, 0xff}).Encode())
            }
            for _, y := range xrange(pmax, -pmax, -pstep) {
                w.Write(etherdream.NewPoint(pmax, y, color.RGBA{0x00, 0xff, 0x00, 0xff}).Encode())
            }
            for _, x := range xrange(pmax, -pmax, -pstep) {
                w.Write(etherdream.NewPoint(x, -pmax, color.RGBA{0x00, 0x00, 0xff, 0xff}).Encode())
            }
            for _, y := range xrange(-pmax, pmax, pstep) {
                w.Write(etherdream.NewPoint(-pmax, y, color.RGBA{0xff, 0xff, 0xff, 0xff}).Encode())
            }
        }
    }

    func xrange(min, max, step int) []int {
        rng := max - min
        ret := make([]int, rng/step+1)
        iY := 0
        for iX := min; rlogic(min, max, iX); iX += step {
            ret[iY] = iX
            iY++
        }
        return ret
    }

    func rlogic(min, max, iX int) bool {
        if min < max {
            return iX <= max
        }
        return iX >= max
    }

## Resources

- [Library Documentation](https://godoc.org/github.com/tgreiser/etherdream)
- [Ether Dream](http://ether-dream.com)
