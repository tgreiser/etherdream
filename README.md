# Go Ether Dream

Go language interface to the Ether Dream laser DAC. Current features: blanking, basic path optimization, quality (trade resolution for frame rate) and 3D scene rendering via [ln](https://github.com/tgreiser/ln). For an introduction to laser projectors and programming, see: [Laser Hack 101 presentation slides](https://github.com/tgreiser/etherdream-touch-designer/blob/master/laser_hack_101.pdf).

Based on the work of [j4cbo](https://github.com/j4cbo/j4cDAC/), [echelon](https://github.com/echelon) and [fogleman](https://github.com/fogleman)

![Spiral](http://prim8.net/art/spiral.jpg)

## Terminology

- projector/scanner - An ILDA compatible laser projector aka laser scanner. 2 (or more) mirrors, 2 galvos, some laser diodes and electronics. Not to be confused with the other kind of laser projector.
- DAC - Digital to Analog Converter. An electronic box that translates digital signals from a computer into analog signals that control the galvos via an IDLA cable. There are proprietary and open source DACs, as well as modified sound cards used as DACs. In this context, it means your Ether Dream(s).

## Setup

This assumes you are plugged in to your ether dream via ethernet cable. You
may need to set up some rules for your firewall. Inbound communications 
are needed for the initial broadcast signal and handshake, if you don't
need to Find a DAC, you can use outbound only.

- Outbound rule: TCP port 7765
- Inbound rule: UDP port 7654

The simplest setup involves one DAC and one projector, but there are many options.

- Multiple projectors chained off one DAC signal
- Multiple projectors chained off one DAC signal, including use of a cross-over ILDA cable to mirror left/right on one side of the room.
- Multiple projectors each with their own DAC. This offers independent control over each projector. If calibrated in a stack, can be used to create complex imagery.

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
    }

## Point Streams

Point streams should contain an infinite loop that will use the [WriteCloser](https://golang.org/pkg/io/#WriteCloser) interface to output encoded points to the DAC sequentially. In Ether Dream, a point has 2D vector information and a color (see: [image/color](https://golang.org/pkg/image/color/#Color)).

    // make a red point at X=0, Y=300
    pt := etherdream.NewPoint(0, 300, color.RGBA{0xff, 0x00, 0x00, 0xff})
    
    // Encode the point to bytes
    by := pt.Encode()
    
    // Stream the encoded points to the DAC
    w.Write(by)
    
From examples\square\square.go:

    func main() {
        ...
        
        debug := false
        dac.Play(squarePointStream, debug)
    }
    func squarePointStream(w io.WriteCloser) {
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

## Blanking and Paths

Here we introduce the use of [tgreiser/ln](https://github.com/tgreiser/ln), a fork of Fogleman's excellent [ln](https://github.com/fogleman/ln) 3D vector library. Blanking is used to reposition the laser to a new location, it involves turning off the beam, repositioning and then a pause. The exact pause necessary to clean up an image can vary from projector to projector so this can be easily configured. I am using the methodology outlined in [Accurate and Efficient Drawing Method for Laser Projection](http://www.art-science.org/journal/v7n4/v7n4pp155/artsci-v7n4pp155.pdf)

If you just want to configure your projector, use examples\parallel_lines\lines.go

    go run examples\parallel_lines\lines.go -pre-blank-count=1 -post-blank-count=5
    # Without sufficient post-blank-count, it produce diagonal lines that cut across most of the image.
    
    go run examples\parallel_lines\lines.go -pre-blank-count=0 -post-blank-count=17
    # These settings look good on my 30 KPPS projectors
    
In the code itself, the flags can be set via etherdream.PreBlankCount and etherdream.PostBlankCount. There is another important variable for the drawing engine, DrawSpeed, which we'll talk about a little later.

    // declare some ln Paths
    p := ln.Path{ln.Vector{0, 0, 0}, ln.Vector{0, 500, 0}}
    p2 := ln.Path{ln.Vector{10000, 0, 0}, ln.Vector{10000, 500, 0}}
    // draw speed 0 will use defaults
    speed := 0
    
    // in the draw loop
    for {
        // draw the first path
        etherdream.DrawPath(w, p, c, speed)
        // use ln Vector.Distance to see if a blank is necessary
        if p2[0].Distance(p[1]) > 0 {
            // blank from p endpoint to p2 startpoint
            etherdream.BlankPath(w, ln.Path{p[1], p2[0]})
        }
        // draw p2
        etherdream.DrawPath(w, p2, c, speed)
        if p2[1].Distance(p[0]) > 0 {
            blank from p2 endbpoint back to original start
            etherdream.BlankPath(w, ln.Path{p2[1], p[0]})
        }
    }

## 3D Rendering

ln can also help you with 3D rendering and transformation. You can position 3D primitives within a scene, render those to paths, optimize the order of the paths and then send the result to the projector. See examples\ln1\ln1.go. Aside from the base ln functionality, the one thing to be aware of here is paths.Optimize - without it the ln output creates many unnessesary blank lines.

    // render our scene to paths
    paths := scene.Render(eye, center, up, width, height, fovy, znear, zfar, step)
    // reorder the paths for optimized output
	paths.Optimize()

    // now we can draw all our paths with the laser

## Draw Speed

When a frame takes too long to draw you will see the output flicker. We can adjust the amount of time we take to draw a path to trade precision for frame rate. This gives you a little more control over the perceived quality of your laser output.

    go run examples\ln2\ln2.go
    # the default draw speed 50 doesn't look very good. Severe flicker.
    
    go run examples\ln2\ln2.go -draw-speed 80
    # when I increase the draw speed some distortion appears on the corners, but flicker is almost entirely eliminated.

## TODO

- A setting for adaptive draw speed.
- Optimization - slow down prior to to sharp angles of movement.
- Import of SVG/ILDA files.

## Resources

- [Library Documentation](https://godoc.org/github.com/tgreiser/etherdream)
- [Ether Dream](http://ether-dream.com)
