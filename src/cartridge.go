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

func (cart *Cartridge) read(address uint16) uint8 {
	return (cart.memory)[address]
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

	cart := new(Cartridge)
	cart.memory = memory
	return cart
}
