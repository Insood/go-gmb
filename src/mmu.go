package main

import (
    "fmt"
    //"github.com/hajimehoshi/ebiten/ebitenutil"
)

// MMU - Memory management unit. Exposes a read/write interface to some internal memory
type MMU struct {
    internalRAM []uint8
    cart *Cartridge
}

// Returns an 8-bit value at the given address
func (mmu *MMU) read8(address uint16) uint8 {
    if address == 0xFF00 { // P1 (joy pad info)
        return 0 // Unimplemented
    } else if address == 0xFF01  { // Serial transfer data
        panic("Reads from 0xFF01 unimplemented")
    } else if address == 0xFF02 { // SC control
        panic("Reads from 0xFF02 unimplemented")
    } else if address == 0xFF41 { 
        panic("Reads from 0xFF41 unimplemented")
    } else if address == 0xFF47 {
        panic("Reads from 0xFF47 unimplemented")
    } else if address == 0xFF48 {
        panic("Reads from 0xFF48 unimplemented")
    } else if address == 0xFF49 {
        panic("Reads from 0xFF49 unimplemented")
    } else if (address >= 0xFF00) && (address <= 0xFFFF) {
        return (mmu.internalRAM)[address]
    }

    return (mmu.cart.memory)[address]
}

// Returns a 16-bit value starting from the given address
// The value returned is formed by: <*address> | <*address+1> << 8
func (mmu *MMU) read16(address uint16) uint16 {
    // TODO: Fix this
    return uint16(mmu.read8(address)) | uint16(mmu.read8(address+1)) << 8
}

// Writes an 8-bit value to the 16-bit address provided.
// TODO: Check to make sure that data is being written to RAM and not ROM
func (mmu *MMU) write8(address uint16, data uint8) {
    if address == 0xFF01 { // Writing to the serial port; used by the test ROM to give output
        fmt.Printf("%c", data) // Now printing debug messages properly
    } else if address == 0xFF02 {
        // SB - do nothing; will not handle
    } else if address == 0xFF04 {
        mmu.internalRAM[0xFF04] = 0 // Increment the DIV (divider register) always resets it to 0
    } else if address == 0xFF44 {
        mmu.internalRAM[0xFF44] = 0 // Incrementing LY (LCDC ycoordinate) always reset it to zero
    } else if address == 0xFF45 {
        panic("0xFF45 unimplemented")
    } else if address == 0xFF46 {
        panic("0xFF46 unimplemented")
    } else if address == 0xFF48 {
        panic("0xFF48 unimplemented")
    } else if address == 0xFF49 {
        panic("0xFF49 unimplemented")
    } else if (address >= 0xFF00) && (address <= 0xFFFF) {
        mmu.internalRAM[address] = data
    } else {
        mmu.cart.memory[address] = data
    }
}

// Writes a 16-bit value to the 16-bit address provided
// The low byte of data is stored at (address)
// The high byte of data is stored at (address+1)
func (mmu *MMU) write16(address uint16, data uint16) {
    mmu.write8(address,uint8(data & 0xFF))
    mmu.write8(address+1,uint8(data >> 8))
}

// incrementDIV - Increment the divider register
// This register cannot be written to normally (writing to it resets it)
// It is only ever incremented by the Timer and only by 1
func (mmu * MMU) incrementDIV() {
    mmu.internalRAM[0xFF04]++
}

// getTIMA - Returns the value of the 8-bit timer register
func (mmu * MMU) getTIMA() uint8{
    return mmu.read8(0xFF05)
}

func (mmu * MMU) setTIMA(newValue uint8) {
    mmu.write8(0xFF05,newValue)
}

// getTMA - returns the timer modulator
// This is the value that TIMA is set to for every overflow
func (mmu * MMU) getTMA() uint8{
    return mmu.read8(0xFF06)
}

// getTAC() - Returns the value inside the timer control register
func (mmu * MMU) getTAC() uint8 {
    return mmu.read8(0xFF07)
}

// getIF() - Returns the value of the interrupt flag
func (mmu * MMU) getIF() uint8{
    return mmu.read8(0xFF0F)
}

// setIF() - Sets the interrupt flag to new values
// This may trigger an interrupt routine during the next instruction
func (mmu * MMU) setIF(newValue uint8){
    mmu.write8(0xFF0F,newValue)
}

// getIE() - Returns the value of the interrupt enabled register
func (mmu * MMU) getIE() uint8 {
    return mmu.read8(0xFFFF)
}

// showDisplay - Returns true if the display should be shown
func (mmu * MMU) showDisplay() bool {
    return (mmu.read8(0xFF40) >> 7) == 0x1
}

// bgTileDataAddress - Returns the address of the given tileNumber
// based on which tileData region is selected in LCDC
func (mmu * MMU) bgTileDataAddress(tileNumber uint8) uint16 {
    tileAddress := uint16(0)
    if ((mmu.read8(0xFF40) >> 4) & 0x1) == 0x1 {
        tileAddress = 0x8000
    } else {
        tileAddress = 0x8800
    }
    return tileAddress + uint16(tileNumber)*16
}

// bgTileMapStartAddress - Returns the start of 1024-byte area which
// contains 32x32 tilemap to use
func (mmu *MMU) bgTileMapStartAddress() uint16 {
    if ((mmu.read8(0xFF40) >> 3) & 0x1) == 0x1 {
        return 0x9C00
    }
    return 0x9800
}

func (mmu * MMU) scrollY() uint8 {
    return mmu.read8(0xFF42)
}

func (mmu * MMU) scrollX() uint8 {
    return mmu.read8(0xFF43)
}

func (mmu * MMU) windowY() uint8 {
    return mmu.read8(0xFF4A)
}

func (mmu *MMU) windowX() uint8 {
    return mmu.read8(0xFF4B)
}

// getLY - Returns the value of the LY register (LCD Y; aka current scanline)
func (mmu * MMU) getLY() uint8 {
    return mmu.read8(0xFF44)
}

// incrementLY - Handles incrementing the LCD Y-Register & setting interrupts
func (mmu * MMU) incrementLY() {
    currentScanline := mmu.read8(0xFF44)
    currentScanline++
    if currentScanline == 144 {
        mmu.setIF(mmu.getIF() | 0x1) // Trigger vblank!
        mmu.internalRAM[0xFF44] = currentScanline
    } else if currentScanline > 153 { // Max number of scanlines have been reached
        mmu.internalRAM[0xFF44] = 0
    } else {
        mmu.internalRAM[0xFF44] = currentScanline
    }
}

// backgroundPixelAt(x,y)
// x,y are coordinates in the BG tile space. To read the interleaved pixel color, do the following:
//  1) Calculate which tile the pixel is in. The tilespace is 32x32 8px tiles in size
//  2) Calculate the address where the tile starts in memory
//  3) Calculate the exact byte which contains the pixel data
//  4) Calculate which bit in the two bytes contain the pixel data
//       
//       tileAddress [0L][1L][2L][3L][4L][5L][6L][7L] // Two bytes contain the
//           +1      [0H][1H][2H][3H][4H][5H][6H][7H] // color data for 8 pixels
//          ...
//          +14      [         last two bytes       ]
//          +15      [        last eight pixels     ]
func (mmu * MMU) backgroundPixelAt(x uint8, y uint8) int{
    // 32 tiles per row. y>>3 (same as y/8) gets the row. x>>3 (x/8) gets the columns
    tileMapOffset := (uint16(x)>>3) + (uint16(y)>>3)*32
    tileSelectionAddress := mmu.bgTileMapStartAddress() + uint16(tileMapOffset)
    tileNumber := mmu.read8(tileSelectionAddress) // Which one of 256 tiles are to be shown
    tileDataAddress := mmu.bgTileDataAddress(tileNumber) // Where the 16-bytes of the tile begin

    tileYOffset := (y & 0x7)*2 // Each row in the tile takes 2 bytes
    tileXOffset := (x & 0x7)   // Each col in the tile is 1 bit
    pixelByte  := tileDataAddress + uint16(tileYOffset)
    pixLow := (mmu.read8(pixelByte+1) >> (7-tileXOffset)) & 0x1
    pixHigh := (mmu.read8(pixelByte) >> (7-tileXOffset)) & 0x1
    colorNumber := (pixHigh << 1) | pixLow
    return GameBoyColorMap[colorNumber]
}

func createMMU() *MMU {
    mmu := new(MMU)
    mmu.internalRAM = make([]uint8, 65536) // Pre-allocate all that beautiful unused memory
    return mmu
}
