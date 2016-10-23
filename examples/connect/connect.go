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

package main

import (
	"log"

	"github.com/tgreiser/etherdream"
)

func main() {
	log.Printf("Listening...\n")
	addr, bp, err := etherdream.FindFirstDAC()
	if err != nil {
		log.Printf("Network error: %v", err)
	}

	log.Printf("Found DAC at %v\n", addr)

	log.Printf("BP:\n%v\n", bp)
	log.Printf("Status:\n%v\n", bp.Status)

	dac, err := etherdream.NewDAC(addr.IP.String())
	if err != nil {
		log.Fatal(err)
	}
	defer dac.Close()
	log.Printf("Initialized %v\n", dac.LastStatus)
	log.Printf("Firmware String: %v\n", dac.FirmwareString)

	st, err := dac.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Ping status: %v", st)
}
