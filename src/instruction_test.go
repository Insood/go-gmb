package main

import "testing"

func testCPU() *CPU {
    cpu := newCPU()

    emptyMemory := make([]uint8, 65536) // Make sure that we have a full 64KB of memory
    cart := new(Cartridge)
    cart.memory = emptyMemory
    cpu.mmu.cart = cart
    return cpu
}
func TestPopHL(t *testing.T) {
    cpu := testCPU()
    cpu.stackPointer = 0x1000
    cpu.mmu.write8(0x1000, 0x55)
    cpu.mmu.write8(0x1001, 0x33)
    cpu.programCounter = 0x100
    cpu.mmu.write8(0x100, 0xE1)
    pop(cpu)

    if cpu.stackPointer != 0x1002 {
        t.Errorf("POP HL: Stack Pointer was not incremented (+2)")
    }

    if cpu.rh != 0x33 {
        t.Errorf("POP HL: rH was not properly set")
    }

    if cpu.rl != 0x55 {
        t.Errorf("POP HL: rL was not properly set")
    }

    if cpu.programCounter != 0x101 {
        t.Errorf("POP HL: Program counter was not properly incremented (+1)")
    }
}

func TestPopAF(t *testing.T) {
    cpu := testCPU()
    cpu.stackPointer = 0x1000
    cpu.mmu.write8(0x1000, 0xA0) // F - zero & half carry are set
    cpu.mmu.write8(0x1001, 0x33) // rA value
    cpu.programCounter = 0x100
    cpu.mmu.write8(0x100, 0xF1) // POP AF
    pop(cpu)
    if cpu.ra != 0x33 {
        t.Errorf("POP AF: rA was not properly set")
    }

    if !cpu.zero {
        t.Errorf("POP AF: Z-flag was not set")
    }

    if cpu.subtract {
        t.Errorf("POP AF: N-flag was set when it shouldn't have been")
    }
    if !cpu.halfCarry {
        t.Errorf("POP AF: H-flag was not set when it should've been")
    }
    if cpu.carry {
        t.Errorf("POP AF: C-flag was set when it shouldn't have been")
    }
    if cpu.programCounter != 0x101 {
        t.Errorf("POP AF: Program counter was not properly incremented (+1)")
    }
    if cpu.stackPointer != 0x1002 {
        t.Errorf("POP AF: Stack Pointer was not incremented (+2)")
    }
}

func TestPush(t *testing.T) {
    cpu := testCPU()
    cpu.rh = 0x22
    cpu.rl = 0x33
    cpu.stackPointer = 0x1007
    cpu.programCounter = 0x100
    cpu.mmu.write8(0x100, 0xE5)
    push(cpu) // Will execute PUSH HL

    if cpu.stackPointer != 0x1005 {
        t.Errorf("PUSH HL: Stack Pointer was not decremented (-2)")
    }
    if cpu.mmu.read8(0x1006) != 0x22 {
        t.Errorf("PUSH HL: rH was not properly stored on the stack")
    }
    if cpu.mmu.read8(0x1005) != 0x33 {
        t.Errorf("PUSH HL: rL was not properly stored on the stack")
    }
    if cpu.programCounter != 0x101 {
        t.Errorf("PUSH HL: Program counter was not properly incremented (+1)")
    }
}
