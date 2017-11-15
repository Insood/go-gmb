package main

import "fmt"

// Instruction - a struct which encapsulates a function pointer and also some information
// about the CPU instruction
type Instruction struct {
	name     string
	dataSize int
	function func(cpu *CPU)
	cycles   int
}

// CPU - Represents the LR35902 CPU
type CPU struct {
	rb, rc, rd, re, rh, rl, ra uint8 // Seven working registers
	rarray                     []*uint8
	programCounter             uint16
	stackPointer               uint16
	mainInstructions           [256]Instruction
	extendedInstructions       [256]Instruction

	cart *Cartridge

	zero      bool
	subtract  bool
	carry     bool
	halfCarry bool
	inte      bool // Whether or not interrupts are enabled

	// The following are not part of the microcontroller spec, but are here to help
	// with the emulation
	instructionsExecuted uint64
	cpuCycles            uint64
}

func (cpu *CPU) pswByte() uint8 {
	var data uint8 = 0x0
	if cpu.zero {
		data |= (0x1 << 7)
	}
	if cpu.subtract {
		data |= (0x1 << 6)
	}
	if cpu.halfCarry {
		data |= (0x1 << 5)
	}
	if cpu.carry {
		data |= (0x1 << 4)
	}

	return data
}

func (cpu *CPU) initializeMainInstructionSet() {
	cpu.mainInstructions[0xCD] = Instruction{"CALL", 3, call, 24}

	cpu.mainInstructions[0xF3] = Instruction{"DI", 1, di, 4}

	cpu.mainInstructions[0x3C] = Instruction{"INC A", 1, inc, 4}
	cpu.mainInstructions[0x04] = Instruction{"INC B", 1, inc, 4}
	cpu.mainInstructions[0x0C] = Instruction{"INC C", 1, inc, 4}
	cpu.mainInstructions[0x14] = Instruction{"INC D", 1, inc, 4}
	cpu.mainInstructions[0x1C] = Instruction{"INC E", 1, inc, 4}
	cpu.mainInstructions[0x24] = Instruction{"INC H", 1, inc, 4}
	cpu.mainInstructions[0x2C] = Instruction{"INC L", 1, inc, 4}
	cpu.mainInstructions[0x34] = Instruction{"INC (HL)", 1, inc, 12}

	cpu.mainInstructions[0xC3] = Instruction{"JP nn", 3, jpnn, 12}
	cpu.mainInstructions[0x18] = Instruction{"JR", 2, jr, 8}
	cpu.mainInstructions[0x20] = Instruction{"JR NZ,r8", 2, jrcc, 12}
	cpu.mainInstructions[0x30] = Instruction{"JR NC,r8", 2, jrcc, 12}
	cpu.mainInstructions[0x28] = Instruction{"JR Z,r8", 2, jrcc, 12}
	cpu.mainInstructions[0x38] = Instruction{"JR C,r8", 2, jrcc, 12}

	cpu.mainInstructions[0x01] = Instruction{"LD BC, d16", 3, ld16, 12}
	cpu.mainInstructions[0x11] = Instruction{"LD DE, d16", 3, ld16, 12}
	cpu.mainInstructions[0x21] = Instruction{"LD HL, d16", 3, ld16, 12}
	cpu.mainInstructions[0x31] = Instruction{"LD SP, d16", 3, ld16, 12}
	cpu.mainInstructions[0x32] = Instruction{"LD (HL-), A", 1, lddHLA, 8}

	cpu.mainInstructions[0x70] = Instruction{"LD (HL) B", 1, ldHLr, 8}
	cpu.mainInstructions[0x71] = Instruction{"LD (HL) C", 1, ldHLr, 8}
	cpu.mainInstructions[0x72] = Instruction{"LD (HL) D", 1, ldHLr, 8}
	cpu.mainInstructions[0x73] = Instruction{"LD (HL) E", 1, ldHLr, 8}
	cpu.mainInstructions[0x74] = Instruction{"LD (HL) H", 1, ldHLr, 8}
	cpu.mainInstructions[0x75] = Instruction{"LD (HL) L", 1, ldHLr, 8}
	// 0x76 is HALT. There is no LD (HL) (HL)
	cpu.mainInstructions[0x77] = Instruction{"LD (HL) A", 1, ldHLr, 8}

	cpu.mainInstructions[0x06] = Instruction{"LD B, d8", 2, ldrn, 8}
	cpu.mainInstructions[0x0E] = Instruction{"LD C, d8", 2, ldrn, 8}
	cpu.mainInstructions[0x16] = Instruction{"LD D, d8", 2, ldrn, 8}
	cpu.mainInstructions[0x1E] = Instruction{"LD E, d8", 2, ldrn, 8}
	cpu.mainInstructions[0x26] = Instruction{"LD H, d8", 2, ldrn, 8}
	cpu.mainInstructions[0x2E] = Instruction{"LD L, d8", 2, ldrn, 8}
	cpu.mainInstructions[0x36] = Instruction{"LD (HL), d8", 2, ldrn, 12}
	cpu.mainInstructions[0x3E] = Instruction{"LD A, d8", 2, ldrn, 8}

	cpu.mainInstructions[0x40] = Instruction{"LD B, B", 1, ldrr, 4}
	cpu.mainInstructions[0x41] = Instruction{"LD B, C", 1, ldrr, 4}
	cpu.mainInstructions[0x42] = Instruction{"LD B, D", 1, ldrr, 4}
	cpu.mainInstructions[0x43] = Instruction{"LD B, E", 1, ldrr, 4}
	cpu.mainInstructions[0x44] = Instruction{"LD B, H", 1, ldrr, 4}
	cpu.mainInstructions[0x45] = Instruction{"LD B, L", 1, ldrr, 4}
	cpu.mainInstructions[0x46] = Instruction{"LD B, (HL)", 1, ldrr, 8}
	cpu.mainInstructions[0x47] = Instruction{"LD B, A", 1, ldrr, 4}

	cpu.mainInstructions[0x48] = Instruction{"LD C, B", 1, ldrr, 4}
	cpu.mainInstructions[0x49] = Instruction{"LD C, C", 1, ldrr, 4}
	cpu.mainInstructions[0x4A] = Instruction{"LD C, D", 1, ldrr, 4}
	cpu.mainInstructions[0x4B] = Instruction{"LD C, E", 1, ldrr, 4}
	cpu.mainInstructions[0x4C] = Instruction{"LD C, H", 1, ldrr, 4}
	cpu.mainInstructions[0x4D] = Instruction{"LD C, L", 1, ldrr, 4}
	cpu.mainInstructions[0x4E] = Instruction{"LD C, (HL)", 1, ldrr, 8}
	cpu.mainInstructions[0x4F] = Instruction{"LD C, A", 1, ldrr, 4}

	cpu.mainInstructions[0x50] = Instruction{"LD D, B", 1, ldrr, 4}
	cpu.mainInstructions[0x51] = Instruction{"LD D, C", 1, ldrr, 4}
	cpu.mainInstructions[0x52] = Instruction{"LD D, D", 1, ldrr, 4}
	cpu.mainInstructions[0x53] = Instruction{"LD D, E", 1, ldrr, 4}
	cpu.mainInstructions[0x54] = Instruction{"LD D, H", 1, ldrr, 4}
	cpu.mainInstructions[0x55] = Instruction{"LD D, L", 1, ldrr, 4}
	cpu.mainInstructions[0x56] = Instruction{"LD D, (HL)", 1, ldrr, 8}
	cpu.mainInstructions[0x57] = Instruction{"LD D, A", 1, ldrr, 4}

	cpu.mainInstructions[0x58] = Instruction{"LD E, B", 1, ldrr, 4}
	cpu.mainInstructions[0x59] = Instruction{"LD E, C", 1, ldrr, 4}
	cpu.mainInstructions[0x5A] = Instruction{"LD E, D", 1, ldrr, 4}
	cpu.mainInstructions[0x5B] = Instruction{"LD E, E", 1, ldrr, 4}
	cpu.mainInstructions[0x5C] = Instruction{"LD E, H", 1, ldrr, 4}
	cpu.mainInstructions[0x5D] = Instruction{"LD E, L", 1, ldrr, 4}
	cpu.mainInstructions[0x5E] = Instruction{"LD E, (HL)", 1, ldrr, 8}
	cpu.mainInstructions[0x5F] = Instruction{"LD E, A", 1, ldrr, 4}

	cpu.mainInstructions[0x60] = Instruction{"LD H, B", 1, ldrr, 4}
	cpu.mainInstructions[0x61] = Instruction{"LD H, C", 1, ldrr, 4}
	cpu.mainInstructions[0x62] = Instruction{"LD H, D", 1, ldrr, 4}
	cpu.mainInstructions[0x63] = Instruction{"LD H, E", 1, ldrr, 4}
	cpu.mainInstructions[0x64] = Instruction{"LD H, H", 1, ldrr, 4}
	cpu.mainInstructions[0x65] = Instruction{"LD H, L", 1, ldrr, 4}
	cpu.mainInstructions[0x66] = Instruction{"LD H, (HL)", 1, ldrr, 8}
	cpu.mainInstructions[0x67] = Instruction{"LD H, A", 1, ldrr, 4}

	cpu.mainInstructions[0x68] = Instruction{"LD L, B", 1, ldrr, 4}
	cpu.mainInstructions[0x69] = Instruction{"LD L, C", 1, ldrr, 4}
	cpu.mainInstructions[0x6A] = Instruction{"LD L, D", 1, ldrr, 4}
	cpu.mainInstructions[0x6B] = Instruction{"LD L, E", 1, ldrr, 4}
	cpu.mainInstructions[0x6C] = Instruction{"LD L, H", 1, ldrr, 4}
	cpu.mainInstructions[0x6D] = Instruction{"LD L, L", 1, ldrr, 4}
	cpu.mainInstructions[0x6E] = Instruction{"LD L, (HL)", 1, ldrr, 8}
	cpu.mainInstructions[0x6F] = Instruction{"LD L, A", 1, ldrr, 4}

	cpu.mainInstructions[0x78] = Instruction{"LD A, B", 1, ldrr, 4}
	cpu.mainInstructions[0x79] = Instruction{"LD A, C", 1, ldrr, 4}
	cpu.mainInstructions[0x7A] = Instruction{"LD A, D", 1, ldrr, 4}
	cpu.mainInstructions[0x7B] = Instruction{"LD A, E", 1, ldrr, 4}
	cpu.mainInstructions[0x7C] = Instruction{"LD A, H", 1, ldrr, 4}
	cpu.mainInstructions[0x7D] = Instruction{"LD A, L", 1, ldrr, 4}
	cpu.mainInstructions[0x7E] = Instruction{"LD A, (HL)", 1, ldrr, 8}
	cpu.mainInstructions[0x7F] = Instruction{"LD A, A", 1, ldrr, 4}

	cpu.mainInstructions[0xE2] = Instruction{"LD (C), A", 1, ldCA, 8}

	cpu.mainInstructions[0xE0] = Instruction{"LDH (n),A", 2, ldhna, 12}
	cpu.mainInstructions[0xEA] = Instruction{"LD (nn), A", 3, ldnna, 16}

	cpu.mainInstructions[0x00] = Instruction{"NOP", 1, nop, 4}

	cpu.mainInstructions[0xC1] = Instruction{"POP BC", 1, pop, 12}
	cpu.mainInstructions[0xD1] = Instruction{"POP DE", 1, pop, 12}
	cpu.mainInstructions[0xE1] = Instruction{"POP HL", 1, pop, 12}
	cpu.mainInstructions[0xF1] = Instruction{"POP AF", 1, pop, 12}

	cpu.mainInstructions[0xC5] = Instruction{"PUSH BC", 1, push, 16}
	cpu.mainInstructions[0xD5] = Instruction{"PUSH DE", 1, push, 16}
	cpu.mainInstructions[0xE5] = Instruction{"PUSH HL", 1, push, 16}
	cpu.mainInstructions[0xF5] = Instruction{"PUSH AF", 1, push, 16}

	cpu.mainInstructions[0xC9] = Instruction{"RET", 1, ret, 16}

	cpu.mainInstructions[0xA8] = Instruction{"XOR B", 1, xor, 4}
	cpu.mainInstructions[0xA9] = Instruction{"XOR C", 1, xor, 4}
	cpu.mainInstructions[0xAA] = Instruction{"XOR D", 1, xor, 4}
	cpu.mainInstructions[0xAB] = Instruction{"XOR E", 1, xor, 4}
	cpu.mainInstructions[0xAC] = Instruction{"XOR H", 1, xor, 4}
	cpu.mainInstructions[0xAD] = Instruction{"XOR L", 1, xor, 4}
	cpu.mainInstructions[0xAE] = Instruction{"XOR HL", 1, xor, 4}
	cpu.mainInstructions[0xAF] = Instruction{"XOR A", 1, xor, 4}

}

func (cpu *CPU) initializeExtendedInstructionSet() {
	// Target register: lowest 3 bits

	// BIT instructions (4x, 5x, 6x, 7x)
	registerNames := [8]string{"B", "C", "D", "E", "H", "L", "(HL)", "A"}

	for i := 0x40; i < 0x80; i++ {
		whichBit := (i >> 3) & 0x7
		registerName := registerNames[i&0x7]

		instructionName := fmt.Sprintf("BIT %d %s", whichBit, registerName)
		if i&0x7 != 6 {
			cpu.extendedInstructions[i] = Instruction{instructionName, 1, bit, 8}
		} else { // Instructions which access (HL) consume twice as many cycles
			cpu.extendedInstructions[i] = Instruction{instructionName, 1, bit, 16}
		}
	}

	// RES instructions (8x, 9x, Ax, Bx)
	// <10XX>

	// SET instructions (Cx, Dx, Ex, Fx)
	// <11XX>

}

func newCPU() *CPU {
	cpu := new(CPU)
	// the 7th element is nil because some instructions have a memory reference
	// bit pattern which corresponds to 110B
	cpu.rarray = []*uint8{&cpu.rb, &cpu.rc, &cpu.rd, &cpu.re, &cpu.rh, &cpu.rl, nil, &cpu.ra}

	for i := 0; i <= 255; i++ {
		cpu.mainInstructions[i] = Instruction{"Unimplemented", 0, unimplemented, 0}
		cpu.extendedInstructions[i] = Instruction{"Unimplemented", 0, unimplemented, 0}
	}

	cpu.initializeMainInstructionSet()
	cpu.initializeExtendedInstructionSet()

	cpu.programCounter = 0x100 // Assuming there's no boot room being executed
	return cpu
}

func (cpu *CPU) getBC() uint16 {
	return uint16(cpu.rb)<<8 | uint16(cpu.rc)
}
func (cpu *CPU) getDE() uint16 {
	return uint16(cpu.rd)<<8 | uint16(cpu.re)
}
func (cpu *CPU) getHL() uint16 {
	return uint16(cpu.rh)<<8 | uint16(cpu.rl)
}
func (cpu *CPU) getAF() uint16 {
	return uint16(cpu.ra)<<8 | uint16(cpu.pswByte())
}

func (cpu *CPU) setBC(data uint16) {
	cpu.rb = uint8(data >> 8)
	cpu.rc = uint8(data & 0xFF)
}
func (cpu *CPU) setDE(data uint16) {
	cpu.rd = uint8(data >> 8)
	cpu.re = uint8(data & 0xFF)
}
func (cpu *CPU) setHL(data uint16) {
	cpu.rh = uint8(data >> 8)
	cpu.rl = uint8(data & 0xFF)
}

// SetRegister - sets the value of a register to the given value
// The register is computed by using the current instruction where
// bits 3,4,5 encode which register pair gets the data
func (cpu *CPU) SetRegister(data uint8) {
	register := (cpu.currentInstruction() >> 3) & 0x7 // 00XXX110
	if register == 0x6 {                              // (HL)
		cpu.SetMemoryReference(data)
	} else {
		*cpu.rarray[register] = data
	}
}

// GetRegisterPair - gets the value of the register pair encoded in the instruction
func (cpu *CPU) GetRegisterPair() uint16 {
	pair := (cpu.currentInstruction() >> 4) & 0x3 // 00XX0000
	switch pair {
	case 0x0:
		return cpu.getBC()
	case 0x1:
		return cpu.getDE()
	case 0x2:
		return cpu.getHL()
	case 0x3:
		return cpu.getAF()
	}
	return 0
}

// SetRegisterPair - sets the value of a register pair to the given value
// The register pair is determined based on the current instruction
// where bits 4/5 encode which register pair gets the data
func (cpu *CPU) SetRegisterPair(data uint16) {
	pair := (cpu.currentInstruction() >> 4) & 0x3 // 00XX0000
	switch pair {
	case 0x0:
		cpu.setBC(data) // Registers B,C
	case 0x1:
		cpu.setDE(data) // Registers D, E
	case 0x2:
		cpu.setHL(data) // Registers H, L
	case 0x3:
		cpu.stackPointer = data
	}
}

// GetMemoryReference - gets the value from the memory specified by registers H & L
func (cpu *CPU) GetMemoryReference() uint8 {
	address := uint16(cpu.rh)<<8 | uint16(cpu.rl)
	return cpu.cart.read8(address)
}

// SetMemoryReference - sets the address stored in (HL) to the given value
func (cpu *CPU) SetMemoryReference(data uint8) {
	address := uint16(cpu.rh)<<8 | uint16(cpu.rl)
	cpu.cart.write8(address, data)
}

// GetRegisterValue - gets the value encoded in the specified register
// Register 6 is the special (HL) register
func (cpu *CPU) GetRegisterValue(register uint8) uint8 {
	if register == 6 {
		return cpu.GetMemoryReference()
	}
	return *cpu.rarray[register]
}

// CheckCondition - checks the condition of the flag encoded in
// bits 3&4 of the CPU instruction and then returns true whether or not
// that condition is met
func (cpu *CPU) CheckCondition() bool {
	condition := (cpu.currentInstruction() >> 3) & 0x3
	var result bool
	switch condition {
	case 0x0:
		result = !cpu.zero // NZ
	case 0x1:
		result = cpu.zero // Z
	case 0x2:
		result = !cpu.carry // NC
	case 0x3:
		result = cpu.carry // C
	}
	return result
}

func (cpu *CPU) currentInstruction() uint8 {
	return cpu.cart.read8(cpu.programCounter)
}
func (cpu *CPU) immediate8() uint8 {
	return cpu.cart.read8(cpu.programCounter + 1)
}
func (cpu *CPU) immediate16() uint16 {
	return cpu.cart.read16(cpu.programCounter + 1)
}

// Sets the Zero bit if bit "b" of the specified register is 0
func bit(cpu *CPU) {
	register := cpu.currentInstruction() & 0x7
	testRegisterValue := cpu.GetRegisterValue(register)
	testBit := (cpu.currentInstruction() >> 3) & 0x7

	cpu.zero = (testRegisterValue>>testBit)&0x1 == 0
	cpu.subtract = false
	cpu.halfCarry = true
	// cpu.carry is not affected by this instruction

	cpu.programCounter++
}

func call(cpu *CPU) {
	target := cpu.immediate16()
	next := cpu.programCounter + 3                        // The instruction after the CALL
	cpu.cart.write8(cpu.stackPointer-2, uint8(next&0xFF)) // LSB
	cpu.cart.write8(cpu.stackPointer-1, uint8(next>>8))   // MSB
	cpu.stackPointer -= 2
	cpu.programCounter = target
}

// di - Disable interrupts
func di(cpu *CPU) {
	cpu.inte = false
	cpu.programCounter++
}

// inc - Increments the value stored in the given register (or memory location)
func inc(cpu *CPU) {
	register := (cpu.currentInstruction() >> 3) & 0x7
	result := cpu.Add(cpu.GetRegisterValue(register), 1, 0)
	cpu.SetRegister(result)
	cpu.programCounter++
}

// jr - jumps relative to the current program counter based on the byte of
// immediate data provided
// NOTE: The immediate byte is a SIGNED value meaning that jumps from
// -126 to +129 are possible

func jr(cpu *CPU) {
	cpu.programCounter = cpu.programCounter + 2 + uint16(int8(cpu.immediate8()))
}

// jrcc - if the specified condition is true, then add the immediate byte
// to the current program counter and then jump to it
// NOTE: The immediate byte is a SIGNED value meaning that jumps from
// -126 to +129 are possible
func jrcc(cpu *CPU) {
	if cpu.CheckCondition() {
		// The jump address is relative to the end of the 2-byte opcode
		cpu.programCounter = cpu.programCounter + 2 + uint16(int8(cpu.immediate8()))
	} else {
		cpu.programCounter += 2
	}
}

// jpnn - jumps to the specified address
func jpnn(cpu *CPU) {
	cpu.programCounter = cpu.immediate16()
}

// Loads the value of register A to the address 0xFF00+C
func ldCA(cpu *CPU) {
	address := uint16(0xFF00) + uint16(cpu.ra)
	cpu.cart.write8(address, cpu.ra)
	cpu.programCounter++
}

// ldrn - Loads 8bit immediate data into the specified register
func ldrn(cpu *CPU) {
	cpu.SetRegister(cpu.immediate8())
	cpu.programCounter += 2
}

// ldrr - Loads register R1 into R2
func ldrr(cpu *CPU) {
	register := cpu.currentInstruction() & 0x7
	value := cpu.GetRegisterValue(register)
	cpu.SetRegister(value)
	cpu.programCounter++
}

// Loads 16-bit immediate data into register pairs
func ld16(cpu *CPU) {
	data16 := cpu.cart.read16(cpu.programCounter + 1)
	cpu.SetRegisterPair(data16)
	cpu.programCounter += 3
}

// ldHLr - Load the contents of register r into (HL)
func ldHLr(cpu *CPU) {
	value := cpu.GetRegisterValue(cpu.currentInstruction() & 0x7)
	cpu.SetMemoryReference(value)
	cpu.programCounter++
}

// ldhna - Loads register A into memory 0xFF00+n
func ldhna(cpu *CPU) {
	target := 0xFF00 + uint16(cpu.immediate8())
	cpu.cart.write8(target, cpu.ra)
	cpu.programCounter += 2
}

// Loads A into the memory address HL, then decrements HL by 1
func lddHLA(cpu *CPU) {
	address := cpu.getHL()
	cpu.cart.write8(address, cpu.ra)
	address--
	cpu.setHL(address)
	cpu.programCounter++
}

// ldnna - loads A into (nn)
func ldnna(cpu *CPU) {
	target := cpu.immediate16()
	cpu.cart.write8(target, cpu.ra)
	cpu.programCounter += 3
}

// nop - do nothing
func nop(cpu *CPU) {
	cpu.programCounter++
}

// pop - moves a value off the top of the stack and into the designated register
// and then increments the stack pointer 2x
func pop(cpu *CPU) {
	value := cpu.cart.read16(cpu.stackPointer)
	target := (cpu.currentInstruction() >> 4) & 0x3

	switch target {
	case 0x0:
		cpu.setBC(value)
	case 0x1:
		cpu.setDE(value)
	case 0x2:
		cpu.setHL(value)
	case 0x3:
		cpu.ra = uint8(value >> 8)
		low := (value & 0xFF)
		cpu.zero = (low>>7)&0x1 == 0x1
		cpu.subtract = (low>>6)&0x1 == 0x1
		cpu.halfCarry = (low>>5)&0x1 == 0x1
		cpu.carry = (low>>4)&0x1 == 0x1

	}
	cpu.programCounter++
	cpu.stackPointer += 2
}

// push - pushes a specified register pair to the stack
func push(cpu *CPU) {
	value := cpu.GetRegisterPair() // returns
	cpu.stackPointer -= 2
	cpu.cart.write16(cpu.stackPointer, value)
	cpu.programCounter++
}

// ret - sets the programCounter to the value currently on the stack
func ret(cpu *CPU) {
	cpu.programCounter = cpu.cart.read16(cpu.stackPointer)
	cpu.stackPointer += 2
}

// xor - Exclusive OR with the accumulator
func xor(cpu *CPU) {
	register := cpu.currentInstruction() & 0x7
	value := cpu.GetRegisterValue(register)
	cpu.ra = cpu.ra ^ value
	cpu.zero = (cpu.ra == 0)
	cpu.halfCarry = false
	cpu.subtract = false
	cpu.carry = false
	cpu.programCounter++
}

func unimplemented(cpu *CPU) {
	panic("This instruction is not implemented")
}

func (cpu *CPU) step() {
	instruction := cpu.currentInstruction()
	instructionInfo := Instruction{}
	if instruction != 0xCB {
		instructionInfo = cpu.mainInstructions[instruction]
	} else {
		cpu.programCounter++
		instruction := cpu.currentInstruction()
		instructionInfo = cpu.extendedInstructions[instruction]
	}

	// The values affected by this instructrion will be shown before the next instruction is
	// executed, but before the debugPrint output is shown
	debugPrint(cpu, instructionInfo.name, instructionInfo.dataSize)
	instructionInfo.function(cpu)
	cpu.instructionsExecuted++
	cpu.cpuCycles += uint64(instructionInfo.cycles)
}
