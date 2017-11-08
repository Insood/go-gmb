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

func pswByte(mc *CPU) uint8 {
	var data uint8 = 0x0
	if mc.zero {
		data |= (0x1 << 7)
	}

	if mc.subtract {
		data |= (0x1 << 6)
	}

	if mc.halfCarry {
		data |= (0x1 << 5)
	}
	if mc.carry {
		data |= (0x1 << 4)
	}

	return data
}

func (cpu *CPU) initializeMainInstructionSet() {
	cpu.mainInstructions[0x00] = Instruction{"NOP", 1, nop, 4}

	cpu.mainInstructions[0x20] = Instruction{"JR NZ,r8", 2, jrcc, 12}
	cpu.mainInstructions[0x30] = Instruction{"JR NC,r8", 2, jrcc, 12}
	cpu.mainInstructions[0x28] = Instruction{"JR Z,r8", 2, jrcc, 12}
	cpu.mainInstructions[0x38] = Instruction{"JR C,r8", 2, jrcc, 12}

	cpu.mainInstructions[0x01] = Instruction{"LD BC, d16", 3, ld16, 12}
	cpu.mainInstructions[0x11] = Instruction{"LD DE, d16", 3, ld16, 12}
	cpu.mainInstructions[0x21] = Instruction{"LD HL, d16", 3, ld16, 12}
	cpu.mainInstructions[0x31] = Instruction{"LD SP, d16", 3, ld16, 12}
	cpu.mainInstructions[0x32] = Instruction{"LD (HL-), A", 1, lddHL, 8}

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

// GetRegisterValue - gets the value of the register encoded in the current
// CPU instruction. These are the 3 lowest bits
func (cpu *CPU) GetRegisterValue() uint8 {
	register := cpu.currentInstruction() & 0x7 // 00000XXX
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

// Sets the Zero bit if bit "b" of the specified register is 0
func bit(cpu *CPU) {
	testRegisterValue := cpu.GetRegisterValue()
	testBit := (cpu.currentInstruction() >> 3) & 0x7

	cpu.zero = (testRegisterValue>>testBit)&0x1 == 0
	cpu.subtract = false
	cpu.halfCarry = true
	// cpu.carry is not affected by this instruction

	cpu.programCounter++
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

// Loads 16-bit immediate data into register pairs
func ld16(cpu *CPU) {
	data16 := cpu.cart.read16(cpu.programCounter + 1)
	cpu.SetRegisterPair(data16)
	cpu.programCounter += 3
}

// Loads A into the memory address HL, then decrements HL by 1
func lddHL(cpu *CPU) {
	address := cpu.getHL()
	cpu.cart.write8(address, cpu.ra)
	address--
	cpu.setHL(address)
	cpu.programCounter++
}

// nop - do nothing
func nop(cpu *CPU) {
	cpu.programCounter++
}

// xor - Exclusive OR with the accumulator
func xor(cpu *CPU) {
	value := cpu.GetRegisterValue()
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

/*
func (mc *microcontroller) data16bit() uint16 {
	// This functions creates a 16-bit value from the low & high bits
	// of the currently active instruction. This is used in many places
	// <instruction> <low bits> <high bits> -> returns (high << 8) | low
	return uint16((*mc.memory)[mc.programCounter+1]) | uint16((*mc.memory)[mc.programCounter+2])<<8
}

func (mc *microcontroller) memoryReference() uint16 {
	// Lots of instructions refer to a memory reference which is the address
	// stored in the H/L registers. The address is (H << 8) & (L)
	// H for high, L for low!
	return (uint16(mc.rh) << 8) | (uint16(mc.rl))
}

// OP-Codes, arranged alphabetically (in the future)

func (mc *microcontroller) aci() {
	// Add immediate to accumulator with carry
	debugPrint(mc, "ACI", 1)
	data := (*mc.memory)[mc.programCounter+1]
	carry := uint8(0)
	if mc.carry {
		carry = 1
	}
	mc.ra = Add(mc.ra, data+carry, mc, 0)
	mc.programCounter += 2
}

func (mc *microcontroller) adc() {
	// Add register or memory to accumulator with carry
	letterMap := string("BCDEHLMA")
	cmd := (*mc.memory)[mc.programCounter] & 0x07
	debugPrint(mc, fmt.Sprintf("ADC %s", string(letterMap[cmd])), 0)
	carry := uint8(0)
	if mc.carry {
		carry = 1
	}
	if cmd == 6 { // Memory reference
		mc.ra = Add(mc.ra, (*mc.memory)[mc.memoryReference()], mc, carry)
	} else {
		mc.ra = Add(mc.ra, *mc.rarray[cmd], mc, carry)
	}
	// for some reason the i8080-core calculates the half-carry
	// flag only as the result of the A+VAL, not as part of A+VAL+C

	mc.programCounter++
}

func (mc *microcontroller) add() {
	letterMap := string("BCDEHLMA")
	cmd := (*mc.memory)[mc.programCounter] & 0x07
	debugPrint(mc, fmt.Sprintf("ADD %s", string(letterMap[cmd])), 0)
	if cmd == 6 { // Memory reference
		mc.ra = Add(mc.ra, (*mc.memory)[mc.memoryReference()], mc, 0)
	} else {
		mc.ra = Add(mc.ra, *mc.rarray[cmd], mc, 0)
	}
	mc.programCounter++
}

func (mc *microcontroller) adi() {
	// ADD immediate to A
	debugPrint(mc, "ADI", 1)
	data := (*mc.memory)[mc.programCounter+1]
	mc.ra = Add(mc.ra, data, mc, 0)
	mc.programCounter += 2
}

func (mc *microcontroller) ana() {
	// AND register or memory w/ accumulator
	letterMap := string("BCDEHLMA")
	cmd := (*mc.memory)[mc.programCounter] & 0x07
	debugPrint(mc, fmt.Sprintf("ANA %s", string(letterMap[cmd])), 0)
	data := uint8(0) // placeholder
	if cmd == 6 {    // Memory location held in HL
		data = (*mc.memory)[mc.memoryReference()]
	} else {
		data = *mc.rarray[cmd]
	}
	//mc.auxCarry is not affected per the 8080 programmer's manual
	// But the 8080/8085 manual states that below is the correct behavior
	// http://bitsavers.trailing-edge.com/pdf/intel/MCS80/9800301D_8080_8085_Assembly_Language_Programming_Manual_May81.pdf
	// pg 1-12
	mc.auxCarry = ((mc.ra | data) & 0x08) != 0
	mc.ra &= data
	mc.carry = false // Per spec, carry bit is always reset
	mc.sign = (mc.ra & 0x80) > 0
	mc.zero = mc.ra == 0
	mc.parity = GetParity(mc.ra)

	mc.programCounter++
}

func (mc *microcontroller) ani() {
	// AND immediate with accumulator
	data := (*mc.memory)[mc.programCounter+1]
	debugPrint(mc, "ANI", 1)
	//mc.auxCarry is not affected per the 8080 programmer's manual
	//but some tests rely on this value to be calculated as follows
	mc.auxCarry = ((mc.ra | data) & 0x08) != 0

	mc.ra = mc.ra & data
	mc.carry = false // Because of the specification
	mc.zero = (mc.ra == 0)
	mc.sign = (mc.ra & 0x80) > 0
	mc.parity = GetParity(mc.ra)

	mc.programCounter += 2
}

func (mc *microcontroller) jc() {
	// Jump if carry
	debugPrint(mc, "JC", 2)
	if mc.carry {
		mc.programCounter = mc.data16bit()
	} else {
		mc.programCounter += 3
	}
}

func (mc *microcontroller) jm() {
	// Jump if sign is 1 (minus)
	debugPrint(mc, "JM", 2)
	if mc.sign {
		mc.programCounter = mc.data16bit()
	} else {
		mc.programCounter += 3
	}
}

func (mc *microcontroller) jmp() {
	// 0xC3: JMP <low bits><high bits> - Set the program counter to the new address
	debugPrint(mc, "JMP", 2)
	mc.programCounter = mc.data16bit()
}

func (mc *microcontroller) jp() {
	// Jump if sign is 0 (plus)
	debugPrint(mc, "JP", 2)
	if mc.sign {
		mc.programCounter += 3
	} else {
		mc.programCounter = mc.data16bit()
	}
}

// jz : Jump if zero bit is 1
func (mc *microcontroller) jz() {
	debugPrint(mc, "JZ", 2)
	if mc.zero {
		mc.programCounter = mc.data16bit()
	} else {
		mc.programCounter += 3
	}
}

// jnz : Jump if zero bit is 0
func (mc *microcontroller) jnz() {
	debugPrint(mc, "JNZ", 2)
	if mc.zero {
		mc.programCounter += 3
	} else {
		mc.programCounter = mc.data16bit()
	}
}

// jnc : Jump if Carry bit is zero
func (mc *microcontroller) jnc() {
	debugPrint(mc, "JNC", 2)
	if mc.carry {
		mc.programCounter += 3
	} else { // No carry so jump
		mc.programCounter = mc.data16bit()
	}
}

// jpe : Jump if Parity bit is one
func (mc *microcontroller) jpe() {
	debugPrint(mc, "JPE", 2)
	if mc.parity {
		mc.programCounter = mc.data16bit()
	} else {
		mc.programCounter += 3
	}
}

// jpo : Jump if Parity bit is zero
func (mc *microcontroller) jpo() {
	debugPrint(mc, "JPO", 2)
	if mc.parity {
		mc.programCounter += 3
	} else {
		mc.programCounter = mc.data16bit()
	}
}

func (mc *microcontroller) lxi() {
	// 0x01, 0x11, 0x21, 0x31 <low data> <high data>
	// Based on the 3rd & 4th most significant bits, set the low/high data
	// To specific registers in memory.
	debugPrint(mc, "LXI", 2)
	target := ((*mc.memory)[mc.programCounter] & 0x30) >> 4
	low := (*mc.memory)[mc.programCounter+1]
	high := (*mc.memory)[mc.programCounter+2]
	switch target {
	case 0x0: // Registers B, C
		mc.rb = high
		mc.rc = low
	case 0x1: // Registers D, E
		mc.rd = high
		mc.re = low
	case 0x2: // Registers H, L
		mc.rh = high
		mc.rl = low
	case 0x3: // Register sp
		mc.stackPointer = mc.data16bit()
	}
	mc.programCounter += 3
}

func (mc *microcontroller) mvi() {
	// (0x06, 0x16, 0x26, 0x36, 0x0E, 0x1E, 0x2E, 0x3E) <data>
	// Sets <data> to the register encoded within the instruction
	debugPrint(mc, "MVI", 1)
	target := ((*mc.memory)[mc.programCounter] & 0x38) >> 3
	data := (*mc.memory)[mc.programCounter+1]
	if target == 6 {
		(*mc.memory)[mc.memoryReference()] = data
	} else {
		*(mc.rarray[target]) = data
	}
	mc.programCounter += 2
}

func (mc *microcontroller) call(silent bool) {
	if !silent {
		debugPrint(mc, "CALL", 2)
	}
	target := mc.data16bit()
	//pcHigh := uint8(mc.stackPointer >> 8)
	//pcLow := uint8(mc.stackPointer & 0xFF)
	next := mc.programCounter + 3                        // The instruction after the CALL
	(*mc.memory)[mc.stackPointer-2] = uint8(next & 0xFF) // LSB
	(*mc.memory)[mc.stackPointer-1] = uint8(next >> 8)   // MSB
	mc.stackPointer -= 2
	mc.programCounter = target
}

func (mc *microcontroller) cc() {
	debugPrint(mc, "CC", 2)
	// Call if Carry bit is 1
	if mc.carry {
		mc.call(true)
	} else {
		mc.programCounter += 3
	}
}

func (mc *microcontroller) cm() {
	debugPrint(mc, "CM", 2)
	// Call if Sign bit is 1
	if mc.sign {
		mc.call(true)
	} else {
		mc.programCounter += 3
	}
}

func (mc *microcontroller) cma() {
	// Complement Accumulator (A = ~A)
	debugPrint(mc, "CMA", 0)
	mc.ra = mc.ra ^ 0xFF
	mc.programCounter++
}

func (mc *microcontroller) cmc() {
	// Complement Carry (carry = !carry)
	debugPrint(mc, "CMC", 0)
	mc.carry = !mc.carry
	mc.programCounter++
}

func (mc *microcontroller) cmp() {
	// Compare accumulator with the given register using subtraction
	// The result is discarded, but the flags are retained
	letterMap := string("BCDEHLMA")
	cmd := (*mc.memory)[mc.programCounter] & 0x07 // Bottom 3 bits

	debugPrint(mc, fmt.Sprintf("CMP %s", string(letterMap[cmd])), 0)
	if cmd == 6 { // Memory reference
		Sub(mc.ra, (*mc.memory)[mc.memoryReference()], mc, 0)
	} else {
		Sub(mc.ra, *mc.rarray[cmd], mc, 0)
	}
	mc.programCounter++
}

func (mc *microcontroller) cnc() {
	// Call if No Carry
	debugPrint(mc, "CNC", 2)
	if mc.carry {
		mc.programCounter += 3
	} else {
		mc.call(true)
	}
}

func (mc *microcontroller) cnz() {
	// Call if Not Zero
	debugPrint(mc, "CNZ", 2)
	if mc.zero {
		mc.programCounter += 3
	} else {
		mc.call(true)
	}
}

func (mc *microcontroller) cp() {
	debugPrint(mc, "CP", 2)
	// Call if Sign bit is 0 (+plus)
	if mc.sign {
		mc.programCounter += 3
	} else {
		mc.call(true)
	}
}

func (mc *microcontroller) cpe() {
	// Call if Parity is Even
	debugPrint(mc, "CPE", 2)
	if mc.parity { // parity==1 is even
		mc.call(true)
	} else {
		mc.programCounter += 3
	}
}
func (mc *microcontroller) cpi() {
	// 0xFE: CPI <data>
	// Compare immediate with accumulator - compares the byte of immediate data
	// with the accumulator using subtraction (A - data) and sets some flags
	debugPrint(mc, "CPI", 1)
	data := (*mc.memory)[mc.programCounter+1]
	Sub(mc.ra, data, mc, 0)
	//fmt.Printf("Comparing %02X to %02X\n", mc.ra, data)
	mc.programCounter += 2
}
func (mc *microcontroller) cpo() {
	// Call if Parity is Odd
	debugPrint(mc, "CPO", 2)
	if mc.parity { // parity==1 is even
		mc.programCounter += 3
	} else {
		mc.call(true)
	}
}

func (mc *microcontroller) cz() {
	// Call if Zero
	debugPrint(mc, "CZ", 2)
	if mc.zero {
		mc.call(true)
	} else {
		mc.programCounter += 3
	}
}

func (mc *microcontroller) daa() {
	debugPrint(mc, "DAA", 0)
	// Decimal adjust accumulator
	carry := mc.carry

	// If the 4LSB of RA are more than 9 or
	// if the aux carry bit is set, increment by 6
	add := uint8(0)
	if (mc.ra&0xF) > 9 || mc.auxCarry {
		add += 0x06
	}
	// Then, take the accumulator and check to see if the 4MSB
	// Are more than 9. If they are, increment by six
	if (((mc.ra >> 4) >= 9) && (mc.ra&0xF > 9)) || carry || (mc.ra>>4) > 9 {
		add += 0x60
		// The specification says that if a carry occured out of the
		// 4MSB, that the carry flag must be set otherwise it is unaffacted
		// and retains the previous value
		carry = true
	}

	mc.ra = Add(mc.ra, add, mc, 0)
	mc.carry = carry // calculated carry value for this op

	mc.programCounter++
}

func (mc *microcontroller) dad() {
	// Double add. This affects carry!
	hl := (uint16(mc.rh) << 8) | (uint16(mc.rl))
	cmd := ((*mc.memory)[mc.programCounter] >> 4) & 0x3
	val := uint16(0)
	switch cmd {
	case 0: // BC
		debugPrint(mc, "DAD BC", 0)
		val = (uint16(mc.rb) << 8) | (uint16(mc.rc))
	case 1: // DE
		debugPrint(mc, "DAD DE", 0)
		val = (uint16(mc.rd) << 8) | (uint16(mc.re))
	case 2: // HL
		debugPrint(mc, "DAD HL", 0)
		val = hl
	case 3: // SP
		debugPrint(mc, "DAD SP", 0)
		val = mc.stackPointer
	}
	result := uint32(hl) + uint32(val)
	mc.rh = uint8(result >> 8)
	mc.rl = uint8(result & 0xFF)
	mc.carry = (result & 0x10000) > 0
	mc.programCounter++
}

func (mc *microcontroller) dcx() {
	// Decrement pair by one
	cmd := ((*mc.memory)[mc.programCounter] >> 4) & 0x3
	val := uint16(0)
	switch cmd {
	case 0: // BC
		debugPrint(mc, "DCX BC", 0)
		val = (uint16(mc.rb) << 8) | (uint16(mc.rc))
		val--
		mc.rb = uint8(val >> 8)
		mc.rc = uint8(val & 0xFF)
	case 1: // DE
		debugPrint(mc, "DCX DE", 0)
		val = (uint16(mc.rd) << 8) | (uint16(mc.re))
		val--
		mc.rd = uint8(val >> 8)
		mc.re = uint8(val & 0xFF)
	case 2: // HL
		debugPrint(mc, "DCX HL", 0)
		val = (uint16(mc.rh) << 8) | (uint16(mc.rl))
		val--
		mc.rh = uint8(val >> 8)
		mc.rl = uint8(val & 0xFF)
	case 3: // SP
		debugPrint(mc, "DCX SP", 0)
		mc.stackPointer--
	default:
		panic("DCX case not processed")
	}

	mc.programCounter++
}

func (mc *microcontroller) di() {
	debugPrint(mc, "DI", 0)
	// Disable interrupt
	mc.inte = false
	mc.programCounter++
}

func (mc *microcontroller) dcr() {
	// Decrement register
	letterMap := string("BCDEHLMA")

	oldCarry := mc.carry // For some reason carry is not affected by DCR
	cmd := ((*mc.memory)[mc.programCounter] >> 3) & 0x07
	debugPrint(mc, fmt.Sprintf("DCR %s", string(letterMap[cmd])), 0)
	if cmd == 6 { // Memory location held in HL
		target := (uint16(mc.rh) << 8) | uint16(mc.rl)
		(*mc.memory)[target] = Sub((*mc.memory)[target], 1, mc, 0)
	} else { // Just decrement the register
		*mc.rarray[cmd] = Sub(*mc.rarray[cmd], 1, mc, 0)
	}
	mc.carry = oldCarry

	mc.programCounter++
}

func (mc *microcontroller) ei() {
	debugPrint(mc, "EI", 0)
	// Enable Interrupt
	mc.inte = true
	mc.programCounter++
}

func (mc *microcontroller) inr() {
	// Increment register
	letterMap := string("BCDEHLMA")
	oldCarry := mc.carry // For some reason INR doesn't affect carry
	cmd := ((*mc.memory)[mc.programCounter] >> 3) & 0x07
	debugPrint(mc, fmt.Sprintf("INR %s", string(letterMap[cmd])), 0)
	if cmd == 6 { // Memory location held in HL
		target := (uint16(mc.rh) << 8) | uint16(mc.rl)
		(*mc.memory)[target] = Add((*mc.memory)[target], 1, mc, 0)
	} else { // Just increment the register
		*mc.rarray[cmd] = Add(*mc.rarray[cmd], 1, mc, 0)
	}
	mc.carry = oldCarry
	mc.programCounter++
}

func (mc *microcontroller) halt() {
	fmt.Printf("halt() is not yet implemented. Caled from %x\n", mc.programCounter)
	panic("Not yet implemented")
}

func (mc *microcontroller) inx() {
	// Increment Register Pair
	// 00: BC, 01: DE, 10: HL, 11: SP
	debugPrint(mc, "INX", 0)
	target := ((*mc.memory)[mc.programCounter] >> 4) & 0x3
	switch target {
	case 0: // BC
		value := ((uint16(mc.rb) << 8) | uint16(mc.rc)) + 1
		mc.rb = uint8(value >> 8)
		mc.rc = uint8(value & 0xFF)
	case 1: // DE
		value := ((uint16(mc.rd) << 8) | uint16(mc.re)) + 1
		mc.rd = uint8(value >> 8)
		mc.re = uint8(value & 0xFF)
	case 2: // HL
		value := ((uint16(mc.rh) << 8) | uint16(mc.rl)) + 1
		mc.rh = uint8(value >> 8)
		mc.rl = uint8(value & 0xFF)
	case 3: // SP
		mc.stackPointer++
	}
	mc.programCounter++
}

func (mc *microcontroller) lda() {
	debugPrint(mc, "LDA", 2)
	// Load Accummulator Direct <low> <high>
	mc.ra = (*mc.memory)[mc.data16bit()]
	mc.programCounter += 3
}

func (mc *microcontroller) ldax() {
	// 0x0A, 0x1A : LDAX (no other data)
	// Load the contents of the memory address either in B/C or D/E
	// into the Accumulator
	debugPrint(mc, "LDAX", 0)
	instruction := ((*mc.memory)[mc.programCounter] >> 4) & 1
	var low, high uint8
	switch instruction {
	case 0x0:
		low = mc.rc  // C
		high = mc.rb // B
	case 0x1:
		low = mc.re  // E
		high = mc.rd // D
	}
	address := (uint16(high) << 8) | uint16(low)
	mc.ra = (*mc.memory)[address]
	mc.programCounter++
}

func (mc *microcontroller) lhld() {
	// Load H&L directly
	debugPrint(mc, "LHLD", 2)
	target := mc.data16bit()
	mc.rl = (*mc.memory)[target]
	mc.rh = (*mc.memory)[target+1]
	mc.programCounter += 3
}

func (mc *microcontroller) mov() {
	letterMap := string("BCDEHLMA")

	dst := ((*mc.memory)[mc.programCounter] >> 3) & 0x7 // Bits 4-6
	src := (*mc.memory)[mc.programCounter] & 0x7        // Lowest 3 bits
	str := fmt.Sprintf("MOV %s%s", string(letterMap[dst]), string(letterMap[src]))
	debugPrint(mc, str, 0)

	var target *uint8
	if dst == 6 { // Memory reference
		target = &(*mc.memory)[mc.memoryReference()] // address of array element
	} else {
		target = mc.rarray[dst]
	}

	var data uint8
	if src == 6 {
		data = (*mc.memory)[mc.memoryReference()]
	} else {
		data = *(mc.rarray[src])
	}
	*target = data

	mc.programCounter++
}
func (mc *microcontroller) nop() {
	debugPrint(mc, "NOP", 0)
	// 0x0: NOP - Do nothing
	// a place to hook in other instructions
	mc.programCounter++
}

func (mc *microcontroller) ora() {
	// OR register or memory w/ accumulator
	letterMap := string("BCDEHLMA")
	cmd := (*mc.memory)[mc.programCounter] & 0x07
	debugPrint(mc, fmt.Sprintf("ORA %s", string(letterMap[cmd])), 0)
	if cmd == 6 { // Memory location held in HL
		target := (uint16(mc.rh) << 8) | uint16(mc.rl)
		mc.ra |= (*mc.memory)[target]
	} else { // Just decrement the register
		mc.ra |= *mc.rarray[cmd]
	}
	mc.carry = false // Per spec, carry bit is always reset
	mc.sign = (mc.ra & 0x80) > 0
	mc.zero = mc.ra == 0
	mc.parity = GetParity(mc.ra)
	// Nothing in spec about mc.auxCarry, but some tests
	// rely on it being reset
	mc.auxCarry = false
	mc.programCounter++
}

func (mc *microcontroller) ori() {
	// OR immediate with accumulator
	data := (*mc.memory)[mc.programCounter+1]
	debugPrint(mc, "ORI", 1)
	mc.ra = mc.ra | data
	mc.carry = false // Because of the specification
	mc.zero = (mc.ra == 0)
	mc.sign = (mc.ra & 0x80) > 0
	mc.parity = GetParity(mc.ra)
	//mc.auxCarry is not affected per the 8080 programmer's manual
	// but some tests rely on it being reset
	mc.auxCarry = false
	mc.programCounter += 2
}

func (mc *microcontroller) pchl() {
	debugPrint(mc, "PCHL", 0)
	low := uint16(mc.rl)
	high := uint16(mc.rh) << 8
	mc.programCounter = high | low
}

func (mc *microcontroller) pop() {
	target := ((*mc.memory)[mc.programCounter] >> 4) & 0x3
	low := (*mc.memory)[mc.stackPointer]
	high := (*mc.memory)[mc.stackPointer+1]
	switch target {
	case 0: // BC
		debugPrint(mc, "POP BC", 0)
		mc.rb = high
		mc.rc = low
	case 1: // DE
		debugPrint(mc, "POP DE", 0)
		mc.rd = high
		mc.re = low
	case 2: // HL
		debugPrint(mc, "POP HL", 0)
		mc.rh = high
		mc.rl = low
	case 3: // flags & A (POP PSW)
		debugPrint(mc, "POP PSW", 0)
		mc.sign = ((low >> 7) & 0x1) == 0x1
		mc.zero = ((low >> 6) & 0x1) == 0x1
		mc.auxCarry = ((low >> 4) & 0x1) == 0x1
		mc.carry = (low & 0x1) == 0x1 // LSB
		mc.parity = ((low >> 2) & 0x1) == 0x1
		mc.ra = high
	}
	mc.programCounter++
	mc.stackPointer += 2
}

func (mc *microcontroller) push() {
	cmd := ((*mc.memory)[mc.programCounter] >> 4) & 0x3
	cmdMap := []string{"BC", "DE", "HL", "PSW"}
	cmdStr := fmt.Sprintf("PUSH %s", cmdMap[cmd])
	debugPrint(mc, cmdStr, 0)
	var first, second uint8
	switch cmd {
	case 0x0: // B & C
		first = mc.rb
		second = mc.rc
	case 0x1: // D & E
		first = mc.rd
		second = mc.re
	case 0x2: // H & L
		first = mc.rh
		second = mc.rl
	case 0x3: // flags & A
		first = mc.ra
		second = pswByte(mc)
	}
	(*mc.memory)[mc.stackPointer-2] = second
	(*mc.memory)[mc.stackPointer-1] = first
	mc.stackPointer -= 2
	mc.programCounter++
}
func (mc *microcontroller) ral() {
	debugPrint(mc, "RAL", 0)
	// Rotate one bit to the left. Highest bit goes to carry
	// Carry becomes LSB
	carry := uint8(0)
	if mc.carry {
		carry = 1
	}
	mc.carry = mc.ra&0x80 > 0 // MSB
	mc.ra = (mc.ra << 1) | carry
	mc.programCounter++
}

func (mc *microcontroller) rar() {
	debugPrint(mc, "RAR", 0)
	// Rotate accumulator to the right by 1 bit
	// Carry becomes the LSB of the accumulator
	// MSB becomes the previous carry value

	carry := uint8(0)
	if mc.carry {
		carry = 1
	}
	mc.carry = mc.ra&0x1 > 0 // LSB
	mc.ra = (mc.ra >> 1) | (carry << 7)

	mc.programCounter++
}

func (mc *microcontroller) ret(silent bool) {
	low := uint16((*mc.memory)[mc.stackPointer])
	high := uint16((*mc.memory)[mc.stackPointer+1])
	target := (high << 8) | low
	if !silent {
		debugPrint(mc, fmt.Sprintf("RET %04X", target), 0)
	}
	mc.programCounter = target
	mc.stackPointer += 2
}

func (mc *microcontroller) retC() {
	// Return if Carry. Called ret_c because there is already an mc.rc
	debugPrint(mc, "RC", 0)
	if mc.carry {
		mc.ret(true)
	} else {
		mc.programCounter++
	}
}

func (mc *microcontroller) rlc() {
	debugPrint(mc, "RLC", 0)
	// Carry bit is set to MSB
	// Rotate accumulator left 1 bit
	// LSB becomes the previous MSB
	msb := mc.ra >> 7

	if msb == 0x1 {
		mc.carry = true
	} else {
		mc.carry = false
	}

	mc.ra = (mc.ra << 1) | msb
	mc.programCounter++
}

func (mc *microcontroller) rm() {
	// Return if Sign bit is 1
	debugPrint(mc, "RM", 0)
	if mc.sign {
		mc.ret(true)
	} else {
		mc.programCounter++
	}
}

func (mc *microcontroller) rnc() {
	// Return it NOT Carry
	debugPrint(mc, "RNC", 0)
	if mc.carry {
		mc.programCounter++
	} else {
		mc.ret(true)
	}
}

func (mc *microcontroller) rnz() {
	// Return it NOT zero
	debugPrint(mc, "RNZ", 0)
	if mc.zero {
		mc.programCounter++
	} else {
		mc.ret(true)
	}
}

func (mc *microcontroller) rp() {
	// Return if Sign bit is 0
	debugPrint(mc, "RP", 0)
	if mc.sign {
		mc.programCounter++
	} else {
		mc.ret(true)
	}
}

func (mc *microcontroller) rpe() {
	// Return if parity is even
	debugPrint(mc, "RPE", 0)
	if mc.parity {
		mc.ret(true)
	} else {
		mc.programCounter++
	}
}

func (mc *microcontroller) rpo() {
	// Return if parity is odd
	debugPrint(mc, "RPO", 0)
	if mc.parity {
		mc.programCounter++
	} else {
		mc.ret(true)
	}
}
func (mc *microcontroller) rrc() {
	debugPrint(mc, "RRC", 0)
	lowBit := mc.ra & 0x1
	// Set the carry bit equal to the LSB
	if lowBit == 0x1 {
		mc.carry = true
	} else {
		mc.carry = false
	}
	mc.ra = (mc.ra >> 1) | (lowBit << 7)
	mc.programCounter++
}
func (mc *microcontroller) rst() {
	// Restart
	debugPrint(mc, "RST", 0)
	exp := ((*mc.memory)[mc.programCounter] >> 3) & 0x7

	(*mc.memory)[mc.stackPointer-2] = uint8(mc.programCounter)      // L
	(*mc.memory)[mc.stackPointer-1] = uint8(mc.programCounter >> 8) // H
	mc.stackPointer -= 2                                            // The manual says (SP) <- (SP)+2, but this is probably wrong

	mc.programCounter = uint16(exp << 3)
}
func (mc *microcontroller) rz() {
	// Return if ZERO
	debugPrint(mc, "RZ", 0)
	if mc.zero {
		mc.ret(true)
	} else {
		mc.programCounter++
	}
}

func (mc *microcontroller) sbi() {
	// Subtract immediate from accumuatlor with borrow
	debugPrint(mc, "SUI", 1)
	carry := uint8(0)
	if mc.carry {
		carry = 1
	}

	data := (*mc.memory)[mc.programCounter+1]
	mc.ra = Sub(mc.ra, data, mc, carry)
	mc.programCounter += 2
}

func (mc *microcontroller) shld() {
	debugPrint(mc, "SHLD", 2)
	// Store H & L directly to memory
	target := mc.data16bit()
	(*mc.memory)[target] = mc.rl
	(*mc.memory)[target+1] = mc.rh
	mc.programCounter += 3
}

func (mc *microcontroller) sphl() {
	debugPrint(mc, "SPHL", 0)
	// SP <- HL
	hl := (uint16(mc.rh) << 8) | (uint16(mc.rl))
	mc.stackPointer = hl
	mc.programCounter++
}

func (mc *microcontroller) sta() {
	// Store accumulator direct at the given address
	debugPrint(mc, "STA", 2)
	(*mc.memory)[mc.data16bit()] = mc.ra
	mc.programCounter += 3
}

func (mc *microcontroller) stax() {
	// 0x02, 0x12 : STAX (no other data)
	// Store the contents of the accumulator at the location pointed to by B/C or D/E
	debugPrint(mc, "STAX", 0)
	instruction := ((*mc.memory)[mc.programCounter] >> 4) & 1
	var low, high uint8
	switch instruction {
	case 0x0:
		low = mc.rc  // C
		high = mc.rb // B
	case 0x1:
		low = mc.re  // E
		high = mc.rd // D
	}
	address := (uint16(high) << 8) | uint16(low)
	(*mc.memory)[address] = mc.ra
	mc.programCounter++
}

func (mc *microcontroller) stc() {
	// Set the carry bit
	debugPrint(mc, "STC", 0)
	mc.carry = true
	mc.programCounter++
}

func (mc *microcontroller) sbb() {
	// Subtract register or memory from accumulator with borrow
	letterMap := string("BCDEHLMA")
	cmd := (*mc.memory)[mc.programCounter] & 0x07 // Bottom 3 bits
	carry := uint8(0)
	if mc.carry {
		carry = 1
	}
	debugPrint(mc, fmt.Sprintf("SBB %s", string(letterMap[cmd])), 0)
	if cmd == 6 { // Memory reference
		mc.ra = Sub(mc.ra, (*mc.memory)[mc.memoryReference()], mc, carry)
	} else {
		mc.ra = Sub(mc.ra, *mc.rarray[cmd], mc, carry)
	}
	mc.programCounter++
}

func (mc *microcontroller) sub() {
	// Subtract based on the register
	letterMap := string("BCDEHLMA")
	cmd := (*mc.memory)[mc.programCounter] & 0x07 // Bottom 3 bits

	debugPrint(mc, fmt.Sprintf("SUB %s", string(letterMap[cmd])), 0)
	if cmd == 6 { // Memory reference
		mc.ra = Sub(mc.ra, (*mc.memory)[mc.memoryReference()], mc, 0)
	} else {
		mc.ra = Sub(mc.ra, *mc.rarray[cmd], mc, 0)
	}
	mc.programCounter++
}
func (mc *microcontroller) sui() {
	// Subtract immediate from accumuatlor
	debugPrint(mc, "SUI", 1)
	data := (*mc.memory)[mc.programCounter+1]
	mc.ra = Sub(mc.ra, data, mc, 0)
	mc.programCounter += 2
}

func (mc *microcontroller) xra() {
	// XOR register or memory w/ accumulator
	letterMap := string("BCDEHLMA")
	cmd := (*mc.memory)[mc.programCounter] & 0x07
	debugPrint(mc, fmt.Sprintf("XRA %s", string(letterMap[cmd])), 0)
	if cmd == 6 { // Memory location held in HL
		target := (uint16(mc.rh) << 8) | uint16(mc.rl)
		mc.ra ^= (*mc.memory)[target]
	} else { // Just decrement the register
		mc.ra ^= *mc.rarray[cmd]
	}
	mc.carry = false // Per spec, carry bit is always reset
	mc.sign = (mc.ra & 0x80) > 0
	mc.zero = mc.ra == 0
	mc.parity = GetParity(mc.ra)
	// Nothing in spec about mc.auxCarry, but some i8080-core
	// tests rely on it being reset
	mc.auxCarry = false
	mc.programCounter++
}

func (mc *microcontroller) xchg() {
	debugPrint(mc, "XCHG", 0)
	// Exchange HL with DE
	h := mc.rh
	l := mc.rl
	mc.rh = mc.rd
	mc.rl = mc.re
	mc.rd = h
	mc.re = l
	mc.programCounter++
}
func (mc *microcontroller) xri() {
	// XOR immediate with accumulator
	data := (*mc.memory)[mc.programCounter+1]
	debugPrint(mc, "XRI", 1)
	mc.ra = mc.ra ^ data
	mc.carry = false // Because of the specification
	mc.zero = (mc.ra == 0)
	mc.sign = (mc.ra & 0x80) > 0
	mc.parity = GetParity(mc.ra)
	//mc.auxCarry is not affected per the 8080 programmer's manual
	// but, some tests rely on it being set to false
	mc.auxCarry = false
	mc.programCounter += 2
}
func (mc *microcontroller) xthl() {
	debugPrint(mc, "XTHL", 0)
	// Exchange stack with values stores in H&L
	low := mc.rl
	high := mc.rh

	mc.rl = (*mc.memory)[mc.stackPointer]
	mc.rh = (*mc.memory)[mc.stackPointer+1]

	(*mc.memory)[mc.stackPointer] = low
	(*mc.memory)[mc.stackPointer+1] = high
	mc.programCounter++
}

func (mc *microcontroller) run() {
	instruction := (*mc.memory)[mc.programCounter]
	switch {
	case instruction == 0xCE:
		mc.aci()
	case (instruction & 0xF8) == 0x88:
		mc.adc()
	case (instruction & 0xF8) == 0x80:
		mc.add()
	case instruction == 0xC6:
		mc.adi()
	case (instruction & 0xF8) == 0xA0:
		mc.ana()
	case instruction == 0xE6:
		mc.ani()
	case instruction == 0xCD:
		mc.call(false)
	case instruction == 0xDC:
		mc.cc()
	case instruction == 0xFC:
		mc.cm()
	case instruction == 0x2F:
		mc.cma()
	case instruction == 0x3F:
		mc.cmc()
	case (instruction & 0xF8) == 0xB8:
		mc.cmp()
	case instruction == 0xD4:
		mc.cnc()
	case instruction == 0xC4:
		mc.cnz()
	case instruction == 0xF4:
		mc.cp()
	case instruction == 0xEC:
		mc.cpe()
	case instruction == 0xFE:
		mc.cpi()
	case instruction == 0xE4:
		mc.cpo()
	case instruction == 0xCC:
		mc.cz()
	case instruction == 0x27:
		mc.daa()
	case (instruction & 0xCF) == 0x09:
		mc.dad()
	case (instruction & 0xCF) == 0x0B:
		mc.dcx()
	case instruction == 0xF3:
		mc.di()
	case (instruction & 0xC7) == 0x05:
		mc.dcr()
	case instruction == 0xFB:
		mc.ei()
	// NOTICE:: Halt MUST be evaulated above MOV because
	// it's similar to a MOV instruction (bit-wise)
	case instruction == 0x76:
		mc.halt()
	case (instruction & 0xC7) == 0x4:
		mc.inr()
	case (instruction & 0xCF) == 0x3: // 0x03, 0x13, 0x23, 0x33
		mc.inx()
	case instruction == 0xC3:
		mc.jmp()
	case instruction == 0xDA:
		mc.jc()
	case instruction == 0xFA:
		mc.jm()
	case instruction == 0xC2:
		mc.jnz()
	case instruction == 0xD2:
		mc.jnc()
	case instruction == 0xF2:
		mc.jp()
	case instruction == 0xEA:
		mc.jpe()
	case instruction == 0xE2:
		mc.jpo()
	case instruction == 0xCA:
		mc.jz()
	case instruction == 0x3A:
		mc.lda()
	case (instruction == 0x0A) || (instruction == 0x1A):
		mc.ldax()
	case instruction == 0x2A:
		mc.lhld()
	case instruction&0xCF == 0x1: //  0x01, 0x11, 0x21, 0x31:
		mc.lxi()
	case (instruction >> 6) == 0x01:
		mc.mov()
	case instruction&0xC7 == 0x6: // 0x06, 0x16, 0x26, 0x36, 0x0E, 0x1E, 0x2E, 0x3E:
		mc.mvi()
	case (instruction & 0xC7) == 0x0: // A bunch of undocumented NOP instructions
		mc.nop()
	case instruction == 0xF6:
		mc.ori()
	case instruction == 0xE9:
		mc.pchl()
	case instruction&0xCF == 0xC1: // 0xC1, 0xD1, 0xE1, 0xF1
		mc.pop()
	case instruction&0xCF == 0xC5: // 0xC5, 0xD5, 0xE5, 0xF5
		mc.push()
	case (instruction & 0xF8) == 0xB0:
		mc.ora()
	case instruction == 0x17:
		mc.ral()
	case instruction == 0x1F:
		mc.rar()
	case instruction == 0xC9:
		mc.ret(false)
	case instruction == 0xD8:
		mc.retC()
	case instruction == 0x07:
		mc.rlc()
	case instruction == 0xF8:
		mc.rm()
	case instruction == 0xD0:
		mc.rnc()
	case instruction == 0xC0:
		mc.rnz()
	case instruction == 0xF0:
		mc.rp()
	case instruction == 0xE8:
		mc.rpe()
	case instruction == 0xE0:
		mc.rpo()
	case instruction == 0x0F:
		mc.rrc()
	case (instruction & 0xC7) == 0xC7:
		mc.rst()
	case instruction == 0xC8:
		mc.rz()
	case instruction == 0xDE:
		mc.sbi()
	case (instruction & 0xF8) == 0x98:
		mc.sbb()
	case (instruction & 0xF8) == 0x90:
		mc.sub()
	case instruction == 0x22:
		mc.shld()
	case instruction == 0xF9:
		mc.sphl()
	case instruction == 0x32:
		mc.sta()
	case instruction == 0x02 || instruction == 0x12:
		mc.stax()
	case instruction == 0x37:
		mc.stc()
	case instruction == 0xD6:
		mc.sui()
	case (instruction & 0xF8) == 0xA8:
		mc.xra()
	case instruction == 0xEB:
		mc.xchg()
	case instruction == 0xEE:
		mc.xri()
	case instruction == 0xE3:
		mc.xthl()
	default:
		err := fmt.Sprintf("[%d] Unknown instruction: %X ", mc.programCounter, instruction)
		fmt.Println(err)
		panic(err)
	}
	mc.instructionsExecuted++
}
*/

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
