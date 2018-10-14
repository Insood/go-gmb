package main

import (
    "fmt"
    //"github.com/hajimehoshi/ebiten/ebitenutil"
)

// MMU - Memory management unit. Exposes a read/write interface to some internal memory
type MMU struct {
    memoryMap []uint8
    cart *Cartridge
}

// Returns an 8-bit value at the given address
func (mmu *MMU) read8(address uint16) uint8 {
    return (mmu.cart.memory)[address]
}

// Returns a 16-bit value starting from the given address
// The value returned is formed by: <*address> | <*address+1> << 8
func (mmu *MMU) read16(address uint16) uint16 {
    return uint16((mmu.cart.memory)[address]) | (uint16((mmu.cart.memory)[address+1]) << 8)
}

// Writes an 8-bit value to the 16-bit address provided.
// TODO: Check to make sure that data is being written to RAM and not ROM
func (mmu *MMU) write8(address uint16, data uint8) {
    mmu.cart.memory[address] = data

    if address == 0xFF01 { // Writing to the serial port; used by the test ROM to give output
        //fmt.Printf("[%X] %c", data, data)
        fmt.Printf("%c", data) // Now printing debug messages properly
    } else if address == 0xFF04 {
        mmu.cart.memory[0xFF04] = 0 // Increment the DIV (divider register) always resets it to 0
    }
}

// Writes a 16-bit value to the 16-bit address provided
// The low byte of data is stored at (address)
// The high byte of data is stored at (address+1)
func (mmu *MMU) write16(address uint16, data uint16) {
    mmu.cart.memory[address] = uint8(data & 0xFF)
    mmu.cart.memory[address+1] = uint8(data >> 8)
}

// incrementDIV - Increment the divider register
// This register cannot be written to normally (writing to it resets it)
// It is only ever incremented by the Timer and only by 1
func (mmu * MMU) incrementDIV() {
    mmu.cart.memory[0xFF04]++
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

func createMMU() *MMU {
    mmu := new(MMU)
    mmu.memoryMap = make([]uint8, 65536) // Pre-allocate all that beautiful memory
    return mmu
}
