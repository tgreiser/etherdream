# Go Ether Dream

Go language interface to the Ether Dream laser DAC. This is a work of progress and currently in a functional alpha state. Check back soon for more examples.

Based on the work of [j4cbo](https://github.com/j4cbo/j4cDAC/)

## Setup

This assumes you are plugged in to your ether dream via ethernet cable. You
may need to set up some rules for your firewall. Inbound communications 
are needed for the initial broadcast signal and handshake, if you don't
need to Find a DAC, you can use outbound only.

- Outbound rule: TCP port 7765
- Inbound rule: UDP port 7654
