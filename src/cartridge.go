package main

import (
	"fmt"
	"io"

	"github.com/hajimehoshi/ebiten/ebitenutil"
)

// Cartridge - exposes a read/write interface to a cartridge memory bank
// based on the current settings
type Cartridge struct {
	memory []uint8
}

// Returns an 8-bit value at the given address
func (cart *Cartridge) read8(address uint16) uint8 {
	return (cart.memory)[address]
}

// Returns a 16-bit value starting from the given address
// The value returned is formed by: <*address> | <*address+1> << 8
func (cart *Cartridge) read16(address uint16) uint16 {
	return uint16((cart.memory)[address]) | (uint16((cart.memory)[address+1]) << 8)
}

// Writes an 8-bit value to the 16-bit address provided.
// TODO: Check to make sure that data is being written to RAM and not ROM
func (cart *Cartridge) write8(address uint16, data uint8) {
	cart.memory[address] = data
}

// loadROM - Reads in the ROM stored in the romname file
// and returns a Cartridge instance that can then be read from/written to
// TODO: Only supports 32KB ROMs (ie: tetris). Implement MBC Type 1+ cartridges
func loadCart(romName string) *Cartridge {
	fi, err := ebitenutil.OpenFile(romName)
	if err != nil {
		fmt.Println(romName, "is an invalid file. Could not open.")
		panic(err)
	}

	memory := make([]uint8, 0, 65536)
	buf := make([]byte, 1024)
	for {
		bytesRead, error := fi.Read(buf)
		slice := buf[0:bytesRead]
		memory = append(memory, slice...) // The ... means to expand the second argument

		if error == io.EOF {
			break
		}
	}

	emptyMemory := make([]uint8, cap(memory)-len(memory)) // Make sure that we have a full 64KB of memory
	memory = append(memory, emptyMemory...)

	cart := new(Cartridge)
	cart.memory = memory
	return cart
}
