package main

// Sub : Subtracts B from A and then sets micro controller flags
// the borrow argument is used by SBC, everything else should call it with a value of 0
func (cpu *CPU) Sub(a uint8, b uint8, borrow uint8) uint8 {
	result16 := uint16(a) - uint16(b) - uint16(borrow)
	result8 := uint8(result16)

	cpu.zero = result8 == 0x0
	cpu.subtract = true
	cpu.carry = result16 > 0xFF
	cpu.halfCarry = (a&0xF - b&0xF - borrow) > 0xF
	return result8
}

// Add : Adds two 1-byte values together and sets the flags
//       the carry flag is used by the ADC instructions
func (cpu *CPU) Add(a uint8, b uint8, carry uint8) uint8 {
	// Do bitwise addition
	result16 := uint16(a) + uint16(b) + uint16(carry)
	result8 := uint8(result16)

	cpu.zero = result8 == 0x0
	cpu.subtract = false
	cpu.carry = result16 > 0xFF
	cpu.halfCarry = ((a & 0xF) + (b & 0xF) + carry) > 0xF
	return result8
}
