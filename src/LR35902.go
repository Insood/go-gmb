package main

import (
    "fmt"
)

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

    //cart *Cartridge
    mmu * MMU;
    timer * Timer;

    halted    bool

    zero      bool
    subtract  bool
    carry     bool
    halfCarry bool
    inte      bool // Whether or not interrupts are enabled

    // The following are not part of the microcontroller spec, but are here to help
    // with the emulation
    instructionsExecuted uint64
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
    cpu.mainInstructions[0x8F] = Instruction{"ADC A,A", 1, adc, 4}
    cpu.mainInstructions[0x88] = Instruction{"ADC A,B", 1, adc, 4}
    cpu.mainInstructions[0x89] = Instruction{"ADC A,C", 1, adc, 4}
    cpu.mainInstructions[0x8A] = Instruction{"ADC A,D", 1, adc, 4}
    cpu.mainInstructions[0x8B] = Instruction{"ADC A,E", 1, adc, 4}
    cpu.mainInstructions[0x8C] = Instruction{"ADC A,H", 1, adc, 4}
    cpu.mainInstructions[0x8D] = Instruction{"ADC A,L", 1, adc, 4}
    cpu.mainInstructions[0x8E] = Instruction{"ADC A,(HL)", 1, adc, 8}
    cpu.mainInstructions[0xCE] = Instruction{"ADC A, d8", 2, adcn, 8}
    cpu.mainInstructions[0x87] = Instruction{"ADD A, A", 1, add, 4}
    cpu.mainInstructions[0x80] = Instruction{"ADD A, B", 1, add, 4}
    cpu.mainInstructions[0x81] = Instruction{"ADD A, C", 1, add, 4}
    cpu.mainInstructions[0x82] = Instruction{"ADD A, D", 1, add, 4}
    cpu.mainInstructions[0x83] = Instruction{"ADD A, E", 1, add, 4}
    cpu.mainInstructions[0x84] = Instruction{"ADD A, H", 1, add, 4}
    cpu.mainInstructions[0x85] = Instruction{"ADD A, L", 1, add, 4}
    cpu.mainInstructions[0x86] = Instruction{"ADD A, (HL)", 1, add, 8}
    cpu.mainInstructions[0xC6] = Instruction{"ADD A, d8", 2, adi, 8}
    cpu.mainInstructions[0x09] = Instruction{"ADD HL, BC", 1, addhl, 8}
    cpu.mainInstructions[0x19] = Instruction{"ADD HL, DE", 1, addhl, 8}
    cpu.mainInstructions[0x29] = Instruction{"ADD HL, HL", 1, addhl, 8}
    cpu.mainInstructions[0x39] = Instruction{"ADD HL, SP", 1, addhl, 8}
    cpu.mainInstructions[0xE8] = Instruction{"ADD SP, n", 2, addspn, 16}
    cpu.mainInstructions[0xA7] = Instruction{"AND A, A", 1, and, 4}
    cpu.mainInstructions[0xA0] = Instruction{"AND A, B", 1, and, 4}
    cpu.mainInstructions[0xA1] = Instruction{"AND A, C", 1, and, 4}
    cpu.mainInstructions[0xA2] = Instruction{"AND A, D", 1, and, 4}
    cpu.mainInstructions[0xA3] = Instruction{"AND A, E", 1, and, 4}
    cpu.mainInstructions[0xA4] = Instruction{"AND A, H", 1, and, 4}
    cpu.mainInstructions[0xA5] = Instruction{"AND A, L", 1, and, 4}
    cpu.mainInstructions[0xA6] = Instruction{"AND A, (HL)", 1, and, 8}
    cpu.mainInstructions[0xE6] = Instruction{"AND d8", 2, ani, 8}

    cpu.mainInstructions[0xCD] = Instruction{"CALL", 3, call, 24}
    cpu.mainInstructions[0xC4] = Instruction{"CALL NZ", 3, callcc, 24}
    cpu.mainInstructions[0xCC] = Instruction{"CALL Z", 3, callcc, 24}
    cpu.mainInstructions[0xD4] = Instruction{"CALL NC", 3, callcc, 24}
    cpu.mainInstructions[0xDC] = Instruction{"CALL C", 3, callcc, 24}
    cpu.mainInstructions[0x3F] = Instruction{"CCF", 1, ccf, 4}
    cpu.mainInstructions[0xBF] = Instruction{"CP A", 1, cpn, 4}
    cpu.mainInstructions[0xB8] = Instruction{"CP B", 1, cpn, 4}
    cpu.mainInstructions[0xB9] = Instruction{"CP C", 1, cpn, 4}
    cpu.mainInstructions[0xBA] = Instruction{"CP D", 1, cpn, 4}
    cpu.mainInstructions[0xBB] = Instruction{"CP E", 1, cpn, 4}
    cpu.mainInstructions[0xBC] = Instruction{"CP H", 1, cpn, 4}
    cpu.mainInstructions[0xBD] = Instruction{"CP L", 1, cpn, 4}
    cpu.mainInstructions[0xBE] = Instruction{"CP (HL)", 1, cpn, 8}
    cpu.mainInstructions[0xFE] = Instruction{"CP d8", 2, cpi, 8}
    cpu.mainInstructions[0x2F] = Instruction{"CPL", 1, cpl, 4}

    cpu.mainInstructions[0x27] = Instruction{"DAA", 1, daa, 4}

    cpu.mainInstructions[0x3D] = Instruction{"DEC A", 1, dec, 4}
    cpu.mainInstructions[0x05] = Instruction{"DEC B", 1, dec, 4}
    cpu.mainInstructions[0x0D] = Instruction{"DEC C", 1, dec, 4}
    cpu.mainInstructions[0x15] = Instruction{"DEC D", 1, dec, 4}
    cpu.mainInstructions[0x1D] = Instruction{"DEC E", 1, dec, 4}
    cpu.mainInstructions[0x25] = Instruction{"DEC H", 1, dec, 4}
    cpu.mainInstructions[0x2D] = Instruction{"DEC L", 1, dec, 4}
    cpu.mainInstructions[0x35] = Instruction{"DEC (HL)", 1, dec, 12}
    cpu.mainInstructions[0x0B] = Instruction{"DEC BC", 1, decrp, 8}
    cpu.mainInstructions[0x1B] = Instruction{"DEC DE", 1, decrp, 8}
    cpu.mainInstructions[0x2B] = Instruction{"DEC HL", 1, decrp, 8}
    cpu.mainInstructions[0x3B] = Instruction{"DEC SP", 1, decrp, 8}
    cpu.mainInstructions[0xF3] = Instruction{"DI", 1, di, 4}
    cpu.mainInstructions[0xFB] = Instruction{"EI", 1, ei, 4}
    cpu.mainInstructions[0x76] = Instruction{"HALT", 1, halt, 4}

    cpu.mainInstructions[0x3C] = Instruction{"INC A", 1, inc, 4}
    cpu.mainInstructions[0x04] = Instruction{"INC B", 1, inc, 4}
    cpu.mainInstructions[0x0C] = Instruction{"INC C", 1, inc, 4}
    cpu.mainInstructions[0x14] = Instruction{"INC D", 1, inc, 4}
    cpu.mainInstructions[0x1C] = Instruction{"INC E", 1, inc, 4}
    cpu.mainInstructions[0x24] = Instruction{"INC H", 1, inc, 4}
    cpu.mainInstructions[0x2C] = Instruction{"INC L", 1, inc, 4}
    cpu.mainInstructions[0x34] = Instruction{"INC (HL)", 1, inc, 12}

    cpu.mainInstructions[0x03] = Instruction{"INC BC", 1, incrp, 8}
    cpu.mainInstructions[0x13] = Instruction{"INC DE", 1, incrp, 8}
    cpu.mainInstructions[0x23] = Instruction{"INC HL", 1, incrp, 8}
    cpu.mainInstructions[0x33] = Instruction{"INC SP", 1, incrp, 8}
    
    cpu.mainInstructions[0xC2] = Instruction{"JP NZ", 3, jpcc, 16}
    cpu.mainInstructions[0xD2] = Instruction{"JP NC", 3, jpcc, 16}
    cpu.mainInstructions[0xCA] = Instruction{"JP Z", 3, jpcc, 16}
    cpu.mainInstructions[0xDA] = Instruction{"JP C", 3, jpcc, 16}
    cpu.mainInstructions[0xE9] = Instruction{"JP (HL)", 1, jphl, 4}
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
    cpu.mainInstructions[0x3A] = Instruction{"LD A, (HL-)", 1, lddAHL, 8}
    cpu.mainInstructions[0x2A] = Instruction{"LD A, (HL+)", 1, ldiAHL, 8}
    cpu.mainInstructions[0x22] = Instruction{"LD (HL+), A", 1, ldiHLA, 8}
    cpu.mainInstructions[0x02] = Instruction{"LD (BC), A", 1, ldBCA, 8}
    cpu.mainInstructions[0x12] = Instruction{"LD (DE), A", 1, ldDEA, 8}
    cpu.mainInstructions[0x70] = Instruction{"LD (HL) B", 1, ldHLr, 8}
    cpu.mainInstructions[0x71] = Instruction{"LD (HL) C", 1, ldHLr, 8}
    cpu.mainInstructions[0x72] = Instruction{"LD (HL) D", 1, ldHLr, 8}
    cpu.mainInstructions[0x73] = Instruction{"LD (HL) E", 1, ldHLr, 8}
    cpu.mainInstructions[0x74] = Instruction{"LD (HL) H", 1, ldHLr, 8}
    cpu.mainInstructions[0x75] = Instruction{"LD (HL) L", 1, ldHLr, 8}
    // 0x76 is HALT. There is no LD (HL) (HL)
    cpu.mainInstructions[0x77] = Instruction{"LD (HL) A", 1, ldHLr, 8}
    cpu.mainInstructions[0xF8] = Instruction{"LD HL, SP+n", 2, ldhlspn, 12}

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
    cpu.mainInstructions[0x0A] = Instruction{"LD A, (BC)", 1, ldabc, 8}
    cpu.mainInstructions[0xFA] = Instruction{"LD A, (nn)", 3, ldann, 16}
    cpu.mainInstructions[0x1A] = Instruction{"LD A, (DE)", 1, ldade, 8}
    cpu.mainInstructions[0xF0] = Instruction{"LD A, ($FF00+n)", 2, ldhan, 12}

    cpu.mainInstructions[0xF2] = Instruction{"LD A, (C)", 1, ldAC, 8}
    cpu.mainInstructions[0xE2] = Instruction{"LD (C), A", 1, ldCA, 8}

    cpu.mainInstructions[0xE0] = Instruction{"LDH (n),A", 2, ldhna, 12}
    cpu.mainInstructions[0xEA] = Instruction{"LD (nn), A", 3, ldnna, 16}
    cpu.mainInstructions[0x08] = Instruction{"LD (nn), SP", 3, ldnnsp, 20}

    cpu.mainInstructions[0xF9] = Instruction{"LD SP, HL", 1, ldsphl, 8}

    cpu.mainInstructions[0x00] = Instruction{"NOP", 1, nop, 4}

    cpu.mainInstructions[0xB0] = Instruction{"OR B", 1, or, 4}
    cpu.mainInstructions[0xB1] = Instruction{"OR C", 1, or, 4}
    cpu.mainInstructions[0xB2] = Instruction{"OR D", 1, or, 4}
    cpu.mainInstructions[0xB3] = Instruction{"OR E", 1, or, 4}
    cpu.mainInstructions[0xB4] = Instruction{"OR H", 1, or, 4}
    cpu.mainInstructions[0xB5] = Instruction{"OR L", 1, or, 4}
    cpu.mainInstructions[0xB6] = Instruction{"OR (HL)", 1, or, 8}
    cpu.mainInstructions[0xB7] = Instruction{"OR A", 1, or, 4}
    cpu.mainInstructions[0xF6] = Instruction{"OR d8", 2, ori, 8}

    cpu.mainInstructions[0xC1] = Instruction{"POP BC", 1, pop, 12}
    cpu.mainInstructions[0xD1] = Instruction{"POP DE", 1, pop, 12}
    cpu.mainInstructions[0xE1] = Instruction{"POP HL", 1, pop, 12}
    cpu.mainInstructions[0xF1] = Instruction{"POP AF", 1, pop, 12}

    cpu.mainInstructions[0xC5] = Instruction{"PUSH BC", 1, push, 16}
    cpu.mainInstructions[0xD5] = Instruction{"PUSH DE", 1, push, 16}
    cpu.mainInstructions[0xE5] = Instruction{"PUSH HL", 1, push, 16}
    cpu.mainInstructions[0xF5] = Instruction{"PUSH AF", 1, push, 16}
    cpu.mainInstructions[0xC9] = Instruction{"RET", 1, ret, 16}
    cpu.mainInstructions[0xC0] = Instruction{"RET NZ", 1, retcc, 8}
    cpu.mainInstructions[0xC8] = Instruction{"RET Z", 1, retcc, 8}
    cpu.mainInstructions[0xD0] = Instruction{"RET NC", 1, retcc, 8}
    cpu.mainInstructions[0xD8] = Instruction{"RET C", 1, retcc, 8}
    cpu.mainInstructions[0xD9] = Instruction{"RETI", 1, reti, 8}
    
    cpu.mainInstructions[0x17] = Instruction{"RLA", 1, rla, 4}
    cpu.mainInstructions[0x07] = Instruction{"RLCA", 1, rlca, 4}
    cpu.mainInstructions[0x1F] = Instruction{"RRA", 1, rra, 4}
    cpu.mainInstructions[0x0F] = Instruction{"RRCA", 1, rrca, 4}

    cpu.mainInstructions[0xC7] = Instruction{"RST 00", 1, rst, 32}
    cpu.mainInstructions[0xCF] = Instruction{"RST 08", 1, rst, 32}
    cpu.mainInstructions[0xD7] = Instruction{"RST 10", 1, rst, 32}
    cpu.mainInstructions[0xDF] = Instruction{"RST 18", 1, rst, 32}
    cpu.mainInstructions[0xE7] = Instruction{"RST 20", 1, rst, 32}
    cpu.mainInstructions[0xEF] = Instruction{"RST 28", 1, rst, 32}
    cpu.mainInstructions[0xF7] = Instruction{"RST 30", 1, rst, 32}
    cpu.mainInstructions[0xFF] = Instruction{"RST 38", 1, rst, 32}
    cpu.mainInstructions[0x9F] = Instruction{"SBC A, A", 1, sbc, 4}
    cpu.mainInstructions[0x98] = Instruction{"SBC A, B", 1, sbc, 4}
    cpu.mainInstructions[0x99] = Instruction{"SBC A, C", 1, sbc, 4}
    cpu.mainInstructions[0x9A] = Instruction{"SBC A, D", 1, sbc, 4}
    cpu.mainInstructions[0x9B] = Instruction{"SBC A, E", 1, sbc, 4}
    cpu.mainInstructions[0x9C] = Instruction{"SBC A, H", 1, sbc, 4}
    cpu.mainInstructions[0x9D] = Instruction{"SBC A, L", 1, sbc, 4}
    cpu.mainInstructions[0x9E] = Instruction{"SBC A, (HL)", 1, sbc, 8}
    cpu.mainInstructions[0xDE] = Instruction{"SBC A, d8", 2, sbcd8, 8}
    cpu.mainInstructions[0x37] = Instruction{"SCF", 1, scf, 4}
    cpu.mainInstructions[0x97] = Instruction{"SUB A, A", 1, sub, 4}
    cpu.mainInstructions[0x90] = Instruction{"SUB A, B", 1, sub, 4}
    cpu.mainInstructions[0x91] = Instruction{"SUB A, C", 1, sub, 4}
    cpu.mainInstructions[0x92] = Instruction{"SUB A, D", 1, sub, 4}
    cpu.mainInstructions[0x93] = Instruction{"SUB A, E", 1, sub, 4}
    cpu.mainInstructions[0x94] = Instruction{"SUB A, H", 1, sub, 4}
    cpu.mainInstructions[0x95] = Instruction{"SUB A, L", 1, sub, 4}
    cpu.mainInstructions[0x96] = Instruction{"SUB A, (HL)", 1, sub, 8}
    cpu.mainInstructions[0xD6] = Instruction{"SUB d8", 2, sbi, 8}

    cpu.mainInstructions[0xA8] = Instruction{"XOR B", 1, xor, 4}
    cpu.mainInstructions[0xA9] = Instruction{"XOR C", 1, xor, 4}
    cpu.mainInstructions[0xAA] = Instruction{"XOR D", 1, xor, 4}
    cpu.mainInstructions[0xAB] = Instruction{"XOR E", 1, xor, 4}
    cpu.mainInstructions[0xAC] = Instruction{"XOR H", 1, xor, 4}
    cpu.mainInstructions[0xAD] = Instruction{"XOR L", 1, xor, 4}
    cpu.mainInstructions[0xAE] = Instruction{"XOR HL", 1, xor, 8}
    cpu.mainInstructions[0xAF] = Instruction{"XOR A", 1, xor, 4}
    cpu.mainInstructions[0xEE] = Instruction{"XOR d8", 2, xord8, 8}

}

func (cpu *CPU) initializeExtendedInstructionSet() {
    cpu.extendedInstructions[0x07] = Instruction{"RLC A", 1, rlc, 8}
    cpu.extendedInstructions[0x00] = Instruction{"RLC B", 1, rlc, 8}
    cpu.extendedInstructions[0x01] = Instruction{"RLC C", 1, rlc, 8}
    cpu.extendedInstructions[0x02] = Instruction{"RLC D", 1, rlc, 8}
    cpu.extendedInstructions[0x03] = Instruction{"RLC E", 1, rlc, 8}
    cpu.extendedInstructions[0x04] = Instruction{"RLC H", 1, rlc, 8}
    cpu.extendedInstructions[0x05] = Instruction{"RLC L", 1, rlc, 8}
    cpu.extendedInstructions[0x06] = Instruction{"RLC (HL)", 1, rlc, 16}

    cpu.extendedInstructions[0x0F] = Instruction{"RRC A", 1, rrc, 8}
    cpu.extendedInstructions[0x08] = Instruction{"RRC B", 1, rrc, 8}
    cpu.extendedInstructions[0x09] = Instruction{"RRC C", 1, rrc, 8}
    cpu.extendedInstructions[0x0A] = Instruction{"RRC D", 1, rrc, 8}
    cpu.extendedInstructions[0x0B] = Instruction{"RRC E", 1, rrc, 8}
    cpu.extendedInstructions[0x0C] = Instruction{"RRC H", 1, rrc, 8}
    cpu.extendedInstructions[0x0D] = Instruction{"RRC L", 1, rrc, 8}
    cpu.extendedInstructions[0x0E] = Instruction{"RRC (HL)", 1, rrc, 16}

    cpu.extendedInstructions[0x17] = Instruction{"RL A", 1, rl, 8}
    cpu.extendedInstructions[0x10] = Instruction{"RL B", 1, rl, 8}
    cpu.extendedInstructions[0x11] = Instruction{"RL C", 1, rl, 8}
    cpu.extendedInstructions[0x12] = Instruction{"RL D", 1, rl, 8}
    cpu.extendedInstructions[0x13] = Instruction{"RL E", 1, rl, 8}
    cpu.extendedInstructions[0x14] = Instruction{"RL H", 1, rl, 8}
    cpu.extendedInstructions[0x15] = Instruction{"RL L", 1, rl, 8}
    cpu.extendedInstructions[0x16] = Instruction{"RL (HL)", 1, rl, 16}

    cpu.extendedInstructions[0x1F] = Instruction{"RRN A", 1, rrn, 8}
    cpu.extendedInstructions[0x18] = Instruction{"RRN B", 1, rrn, 8}
    cpu.extendedInstructions[0x19] = Instruction{"RRN C", 1, rrn, 8}
    cpu.extendedInstructions[0x1A] = Instruction{"RRN D", 1, rrn, 8}
    cpu.extendedInstructions[0x1B] = Instruction{"RRN E", 1, rrn, 8}
    cpu.extendedInstructions[0x1C] = Instruction{"RRN H", 1, rrn, 8}
    cpu.extendedInstructions[0x1D] = Instruction{"RRN L", 1, rrn, 8}
    cpu.extendedInstructions[0x1E] = Instruction{"RRN (HL)", 1, rrn, 16}

    cpu.extendedInstructions[0x27] = Instruction{"SLA A", 1, sla, 8}
    cpu.extendedInstructions[0x20] = Instruction{"SLA B", 1, sla, 8}
    cpu.extendedInstructions[0x21] = Instruction{"SLA C", 1, sla, 8}
    cpu.extendedInstructions[0x22] = Instruction{"SLA D", 1, sla, 8}
    cpu.extendedInstructions[0x23] = Instruction{"SLA E", 1, sla, 8}
    cpu.extendedInstructions[0x24] = Instruction{"SLA H", 1, sla, 8}
    cpu.extendedInstructions[0x25] = Instruction{"SLA L", 1, sla, 8}
    cpu.extendedInstructions[0x26] = Instruction{"SLA (HL)", 1, sla, 16}

    cpu.extendedInstructions[0x2F] = Instruction{"SRA A", 1, sra, 8}
    cpu.extendedInstructions[0x28] = Instruction{"SRA B", 1, sra, 8}
    cpu.extendedInstructions[0x29] = Instruction{"SRA C", 1, sra, 8}
    cpu.extendedInstructions[0x2A] = Instruction{"SRA D", 1, sra, 8}
    cpu.extendedInstructions[0x2B] = Instruction{"SRA E", 1, sra, 8}
    cpu.extendedInstructions[0x2C] = Instruction{"SRA H", 1, sra, 8}
    cpu.extendedInstructions[0x2D] = Instruction{"SRA L", 1, sra, 8}
    cpu.extendedInstructions[0x2E] = Instruction{"SRA (HL)", 1, sra, 16}

    cpu.extendedInstructions[0x3F] = Instruction{"SRL A", 1, srl, 8}
    cpu.extendedInstructions[0x38] = Instruction{"SRL B", 1, srl, 8}
    cpu.extendedInstructions[0x39] = Instruction{"SRL C", 1, srl, 8}
    cpu.extendedInstructions[0x3A] = Instruction{"SRL D", 1, srl, 8}
    cpu.extendedInstructions[0x3B] = Instruction{"SRL E", 1, srl, 8}
    cpu.extendedInstructions[0x3C] = Instruction{"SRL H", 1, srl, 8}
    cpu.extendedInstructions[0x3D] = Instruction{"SRL L", 1, srl, 8}
    cpu.extendedInstructions[0x3E] = Instruction{"SRL (HL)", 1, srl, 16}

    cpu.extendedInstructions[0x37] = Instruction{"SWAP A", 1, swap, 8}
    cpu.extendedInstructions[0x30] = Instruction{"SWAP B", 1, swap, 8}
    cpu.extendedInstructions[0x31] = Instruction{"SWAP C", 1, swap, 8}
    cpu.extendedInstructions[0x32] = Instruction{"SWAP D", 1, swap, 8}
    cpu.extendedInstructions[0x33] = Instruction{"SWAP E", 1, swap, 8}
    cpu.extendedInstructions[0x34] = Instruction{"SWAP H", 1, swap, 8}
    cpu.extendedInstructions[0x35] = Instruction{"SWAP L", 1, swap, 8}
    cpu.extendedInstructions[0x36] = Instruction{"SWAP (HL)", 1, swap, 16}

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
    for i := 0x80; i < 0xC0; i++ {
        whichBit := (i >> 3) & 0x7
        registerName := registerNames[i&0x7]
        instructionName := fmt.Sprintf("RES %d %s", whichBit, registerName)
        if i&0x7 != 6 {
            cpu.extendedInstructions[i] = Instruction{instructionName, 1, res, 8}
        } else { // Instructions which access (HL) consume twice as many cycles
            cpu.extendedInstructions[i] = Instruction{instructionName, 1, res, 16}
        }
    }

    // SET instructions (Cx, Dx, Ex, Fx)
    for i := 0xC0; i <= 0xFF; i++ {
        whichBit := (i >> 3) & 0x7
        registerName := registerNames[i&0x7]
        instructionName := fmt.Sprintf("SET %d %s", whichBit, registerName)
        if i&0x7 != 6 {
            cpu.extendedInstructions[i] = Instruction{instructionName, 1, set, 8}
        } else { // Instructions which access (HL) consume twice as many cycles
            cpu.extendedInstructions[i] = Instruction{instructionName, 1, set, 16}
        }
    }
}

func newCPU() *CPU {
    cpu := new(CPU)
    // the 7th element is nil because some instructions have a memory reference
    // bit pattern which corresponds to 110B
    cpu.rarray = []*uint8{&cpu.rb, &cpu.rc, &cpu.rd, &cpu.re, &cpu.rh, &cpu.rl, nil, &cpu.ra}

    cpu.mmu = createMMU()
    cpu.timer = createTimer(cpu.mmu)

    for i := 0; i <= 255; i++ {
        cpu.mainInstructions[i] = Instruction{"Unimplemented", 0, unimplemented, 0}
        cpu.extendedInstructions[i] = Instruction{"Unimplemented", 0, unimplementedExtended, 0}
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
func (cpu *CPU) SetRegister(register uint8, data uint8) {
    if register == 0x6 { // (HL)
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
    return cpu.mmu.read8(address)
}

// SetMemoryReference - sets the address stored in (HL) to the given value
func (cpu *CPU) SetMemoryReference(data uint8) {
    address := uint16(cpu.rh)<<8 | uint16(cpu.rl)
    cpu.mmu.write8(address, data)
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
    return cpu.mmu.read8(cpu.programCounter)
}
func (cpu *CPU) nextInstruction() uint8 {
    return cpu.mmu.read8(cpu.programCounter+1)
}

func (cpu *CPU) immediate8() uint8 {
    return cpu.mmu.read8(cpu.programCounter + 1)
}
func (cpu *CPU) immediate16() uint16 {
    return cpu.mmu.read16(cpu.programCounter + 1)
}

// adc - add the given register to A with carry
func adc(cpu *CPU) {
    value := cpu.GetRegisterValue(cpu.currentInstruction() & 0x7)
    if cpu.carry {
        cpu.ra = cpu.Add(cpu.ra, value, 1)
    } else {
        cpu.ra = cpu.Add(cpu.ra, value, 0)
    }
    cpu.programCounter++
}

// adc - add immediate value to with carry
func adcn(cpu *CPU) {
    carry := uint8(0)
    if cpu.carry {
        carry = 1
    }
    cpu.ra = cpu.Add(cpu.ra, cpu.immediate8(), carry)
    cpu.programCounter += 2
}

// add - add the value in the given register to A
func add(cpu *CPU) {
    register := cpu.currentInstruction() & 0x7
    value := cpu.GetRegisterValue(register)
    cpu.ra = cpu.Add(cpu.ra, value, 0)
    cpu.programCounter++
}

// addhl - Adds the value of the given register pair (or SP) to HL
// and then sets HL
func addhl(cpu *CPU) {
    target := (cpu.currentInstruction() >> 4) & 0x3
    value := uint16(0)
    // The existing GetRegisterPair() function calls getAF() for case 0x3
    // so let's unfold the function here
    switch target {
    case 0x0:
        value = cpu.getBC()
    case 0x1:
        value = cpu.getDE()
    case 0x2:
        value = cpu.getHL()
    case 0x3:
        value = cpu.stackPointer
    }
    hl := cpu.getHL()
    result32 := uint32(hl) + uint32(value)

    // cpu.zero - not affected
    cpu.subtract = false
    cpu.carry = result32 > 0xFFFF                 // Overflow into 16th bit
    if ((value & 0xFFF) + (hl & 0xFFF)) > 0xFFF { // overflow into 12th bit
        cpu.halfCarry = true
    } else {
        cpu.halfCarry = false
    }
    cpu.setHL(uint16(result32))
    cpu.programCounter++
}

// addspn - Add n to the stack pointer
func addspn(cpu *CPU) {
    // This function is not documented very well in the GameBoy CPU Manual
    // the implementation below is cribbed from the MAME emulator
    n := cpu.immediate8()
    spLower := uint8(cpu.stackPointer & 0xFF)
    cpu.Add(n, spLower, 0) // Set the carry/half flags, but discard the result
    cpu.zero = false       // reset zero
    cpu.subtract = false   // reset subtract
    cpu.stackPointer += uint16(int8(n))
    cpu.programCounter += 2
}

// adi - Adds the immediate value to A
func adi(cpu *CPU) {
    cpu.ra = cpu.Add(cpu.ra, cpu.immediate8(), 0)
    cpu.programCounter += 2
}

// and - perform a logical AND of A with the given register
func and(cpu *CPU) {
    value := cpu.GetRegisterValue(cpu.currentInstruction() & 0x7)
    cpu.ra = cpu.ra & value
    cpu.zero = cpu.ra == 0x0
    cpu.subtract = false
    cpu.halfCarry = true
    cpu.carry = false
    cpu.programCounter++
}

// ani - performs a logical AND of A with the immediate value
func ani(cpu *CPU) {
    result := cpu.immediate8() & cpu.ra
    cpu.ra = result
    cpu.halfCarry = true
    cpu.carry = false
    cpu.subtract = false
    cpu.zero = result == 0
    cpu.programCounter += 2
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
    cpu.mmu.write8(cpu.stackPointer-2, uint8(next&0xFF)) // LSB
    cpu.mmu.write8(cpu.stackPointer-1, uint8(next>>8))   // MSB
    cpu.stackPointer -= 2
    cpu.programCounter = target
}

// callcc - if the specified condition is true, then perform a standard
// call and if not, then just skip over 2 bytes
func callcc(cpu *CPU) {
    if cpu.CheckCondition() {
        call(cpu)
    } else {
        cpu.programCounter += 3
    }
}

// ccf - complement carry flag (!cpu.Carry)
func ccf(cpu *CPU) {
    cpu.carry = !cpu.carry
    // cpu.zero - not affected
    cpu.subtract = false
    cpu.halfCarry = false
    cpu.programCounter++
}

// cpi - Compare A with the immediate value
func cpi(cpu *CPU) {
    value := cpu.immediate8()
    cpu.Sub(cpu.ra, value, 0) // Discard the result, we're only interested in setting the flags
    cpu.programCounter += 2
}

// cpl - Complement A register (bitwise NOT)
func cpl(cpu *CPU) {
    cpu.ra = ^cpu.ra
    // cpu.zero - not affected
    // cpu.carry - not affected
    cpu.subtract = true
    cpu.halfCarry = true
    cpu.programCounter++
}

// cpn - compare A with the given register by doing A-n and throwing away the result
func cpn(cpu *CPU) {
    register := cpu.currentInstruction() & 0x7
    cpu.Sub(cpu.ra, cpu.GetRegisterValue(register), 0)
    cpu.programCounter++
}

// DAA - decimal adjust register A
// Shamelessly implemented based on the notes here: https://ehaskins.com/2018-01-30%20Z80%20DAA/
func daa(cpu *CPU) {
    correction := int16(0)

    if(cpu.halfCarry || (!cpu.subtract && (cpu.ra & 0xF) > 9)){
        correction = 0x06;
    }

    // Checking to see if RA > 0x99 because a value of say.. 9A is invalid (technically 100 in BCD)
    if(cpu.carry || (!cpu.subtract && cpu.ra > 0x99)){
        correction |= 0x60;
        cpu.carry = true;
    }

    a16 := int16(cpu.ra)
    if (!cpu.subtract) {
        a16 += correction
    } else {
        a16 -= correction
    }

    cpu.ra = uint8(a16)
    cpu.zero = cpu.ra == 0
    cpu.halfCarry = false // Always reset
    cpu.programCounter++
}

// dec - decrement the given register by 1 and set some flags
func dec(cpu *CPU) {
    register := (cpu.currentInstruction() >> 3) & 0x7
    value := cpu.GetRegisterValue(register)
    value--
    cpu.SetRegister(register, value)
    //cpu.carry is unaffected
    cpu.subtract = true
    cpu.zero = value == 0
    cpu.halfCarry = (value & 0xF) == 0xF
    cpu.programCounter++
}

// decrp - Decrement the value stored in the register pair
func decrp(cpu *CPU) {
    target := (cpu.currentInstruction() >> 4) & 0x3
    value := uint16(0)
    // The existing GetRegisterPair () function calls getAF() for case 0x3
    // so let's unfold the function here
    switch target {
    case 0x0:
        value = cpu.getBC()
    case 0x1:
        value = cpu.getDE()
    case 0x2:
        value = cpu.getHL()
    case 0x3:
        value = cpu.stackPointer
    }
    value--
    cpu.SetRegisterPair(value)
    cpu.programCounter++
}

// di - Disable interrupts
func di(cpu *CPU) {
    cpu.inte = false
    cpu.programCounter++
}

// ei - enable interrupts
func ei(cpu * CPU){
    cpu.inte = true
    cpu.programCounter++
}
// halt - halts execution until an interrupt fires
// Only has an effect if interrupts have been enabled through EI
// There is a bug with HALT that needs to be properly emulated if
// if it is called with DI disabled
// Lots of helpful information here: https://github.com/AntonioND/giibiiadvance/tree/master/docs
func halt(cpu * CPU){
    if cpu.inte {
        cpu.halted = true
    } else {
        if (cpu.mmu.getIE() & cpu.mmu.getIF()) != 0 {
            // CPU should fail to increase PC and also not clear IF flags
            panic("Unhandled halt bug")
        } else {
            cpu.halted = true
        }
    }
}

// inc - Increments the value stored in the given register (or memory location)
func inc(cpu *CPU) {
    register := (cpu.currentInstruction() >> 3) & 0x7
    oldCarry := cpu.carry
    result := cpu.Add(cpu.GetRegisterValue(register), 1, 0)
    cpu.carry = oldCarry // Carry is not affected by this op
    cpu.SetRegister(register, result)
    cpu.programCounter++
}

// incrp - Increment the value stored in the register pair
func incrp(cpu *CPU) {
    target := (cpu.currentInstruction() >> 4) & 0x3
    value := uint16(0)
    // The existing GetRegisterPair () function calls getAF() for case 0x3
    // so let's unfold the function here
    switch target {
    case 0x0:
        value = cpu.getBC()
    case 0x1:
        value = cpu.getDE()
    case 0x2:
        value = cpu.getHL()
    case 0x3:
        value = cpu.stackPointer
    }
    value++
    cpu.SetRegisterPair(value)
    cpu.programCounter++
}

// jpcc - if the specified condition is true, then perform a jump
// to the specified address
func jpcc(cpu * CPU){
    if cpu.CheckCondition() {
        cpu.programCounter = cpu.immediate16()
    } else {
        cpu.programCounter += 3
    }
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
        /*if int8( cpu.immediate8() )== -2 {
            fmt.Printf("Infinite JR detected")
            os.Exit(-1);
        }*/	
    } else {
        cpu.programCounter += 2
    }
}

// jphl - jump to address in (hl)
func jphl(cpu *CPU) {
    cpu.programCounter = cpu.getHL()
}

// jpnn - jumps to the specified address
func jpnn(cpu *CPU) {
    cpu.programCounter = cpu.immediate16()
}

// ldAC - Load the value in 0xFF00+C into A
func ldAC(cpu *CPU) {
    address := uint16(0xFF00) + uint16(cpu.rc)
    cpu.ra = cpu.mmu.read8(address)
    cpu.programCounter++
}

// lcCA - Loads the value of register A to the address 0xFF00+C
func ldCA(cpu *CPU) {
    address := uint16(0xFF00) + uint16(cpu.rc)
    cpu.mmu.write8(address, cpu.ra)
    cpu.programCounter++
}

// ldhan - (Load high + n into A) Loads the memory in $FF00+n into A
func ldhan(cpu *CPU) {
    address := 0xFF00 + uint16(cpu.immediate8())
    value := cpu.mmu.read8(address)
    cpu.ra = value
    cpu.programCounter += 2
}

// ldrn - Loads 8bit immediate data into the specified register
func ldrn(cpu *CPU) {
    register := (cpu.currentInstruction() >> 3) & 0x7
    cpu.SetRegister(register, cpu.immediate8())
    cpu.programCounter += 2
}

// ldrr - Loads register R1 into R2
func ldrr(cpu *CPU) {
    sourceRegister := cpu.currentInstruction() & 0x7
    targetRegister := (cpu.currentInstruction() >> 3) & 0x7
    value := cpu.GetRegisterValue(sourceRegister)
    cpu.SetRegister(targetRegister, value)
    cpu.programCounter++
}

// Loads 16-bit immediate data into register pairs
func ld16(cpu *CPU) {
    data16 := cpu.mmu.read16(cpu.programCounter + 1)
    cpu.SetRegisterPair(data16)
    cpu.programCounter += 3
}

// ldabc - Loads (bc) into a
func ldabc(cpu *CPU) {
    cpu.ra = cpu.mmu.read8(cpu.getBC())
    cpu.programCounter++
}

// ldade - Loads (de) into a
func ldade(cpu *CPU) {
    cpu.ra = cpu.mmu.read8(cpu.getDE())
    cpu.programCounter++
}

// ldann - Loads (nn) into a
func ldann(cpu *CPU) {
    address := cpu.immediate16()
    value := cpu.mmu.read8(address)
    cpu.ra = value
    cpu.programCounter += 3
}

//ldBCA - Load A into (BC)
func ldBCA(cpu *CPU) {
    address := cpu.getBC()
    cpu.mmu.write8(address, cpu.ra)
    cpu.programCounter++
}

// ldDEA - Load A into (DE)
func ldDEA(cpu *CPU) {
    address := cpu.getDE()
    cpu.mmu.write8(address, cpu.ra)
    cpu.programCounter++
}

// ldHLr - Load the contents of register r into (HL)
func ldHLr(cpu *CPU) {
    value := cpu.GetRegisterValue(cpu.currentInstruction() & 0x7)
    cpu.SetMemoryReference(value)
    cpu.programCounter++
}

// ldhlspn - Load SP+n into HL
func ldhlspn(cpu *CPU) {
    n := cpu.immediate8()
    spLower := uint8(cpu.stackPointer & 0xFF)
    cpu.Add(n, spLower, 0) // Set the carry/half flags, but discard the result
    cpu.zero = false
    cpu.subtract = false
    result := cpu.stackPointer + uint16(int8(n))
    cpu.setHL(result)
    cpu.programCounter += 2
}

// ldhna - Loads register A into memory 0xFF00+n
func ldhna(cpu *CPU) {
    target := 0xFF00 + uint16(cpu.immediate8())
    cpu.mmu.write8(target, cpu.ra)
    cpu.programCounter += 2
}

// lddHLA - Loads A into the memory address HL, then decrements HL by 1
func lddHLA(cpu *CPU) {
    address := cpu.getHL()
    cpu.mmu.write8(address, cpu.ra)
    address--
    cpu.setHL(address)
    cpu.programCounter++
}

// ldiHL - loads A into (HL), then increment HL by 1
func ldiHLA(cpu *CPU) {
    address := cpu.getHL()
    cpu.mmu.write8(address, cpu.ra)
    address++
    cpu.setHL(address)
    cpu.programCounter++
}

// lddAHL - Put (HL) into A, then decrement HL
func lddAHL(cpu *CPU) {
    address := cpu.getHL()
    cpu.ra = cpu.mmu.read8(address)
    address--
    cpu.setHL(address)
    cpu.programCounter++
}

// ldiHLA - Put (hl) into A, then increment HL
// No flags affected
func ldiAHL(cpu *CPU) {
    address := cpu.getHL()
    cpu.ra = cpu.mmu.read8(address)
    address++
    cpu.setHL(address)
    cpu.programCounter++
}

// ldnna - loads A into (nn)
func ldnna(cpu *CPU) {
    target := cpu.immediate16()
    cpu.mmu.write8(target, cpu.ra)
    cpu.programCounter += 3
}

// ldnnsp - Loads the SP into (nn)
func ldnnsp(cpu *CPU) {
    cpu.mmu.write16(cpu.immediate16(), cpu.stackPointer)
    cpu.programCounter += 3
}

// ldsphl - Loads HL into the stack pointer
func ldsphl(cpu *CPU) {
    cpu.stackPointer = cpu.getHL()
    cpu.programCounter++
}

// nop - do nothing
func nop(cpu *CPU) {
    cpu.programCounter++
}

// or - logical or of the specified register with A with the result stored in A
func or(cpu *CPU) {
    register := cpu.currentInstruction() & 0x7
    cpu.ra = cpu.ra | cpu.GetRegisterValue(register)
    cpu.zero = cpu.ra == 0x0
    cpu.subtract = false
    cpu.halfCarry = false
    cpu.carry = false

    cpu.programCounter++
}

// ori - logical or of A with the immediate value
func ori(cpu *CPU) {
    cpu.ra = cpu.ra | cpu.immediate8()
    cpu.zero = cpu.ra == 0x0
    cpu.subtract = false
    cpu.halfCarry = false
    cpu.carry = false
    cpu.programCounter += 2
}

// pop - moves a value off the top of the stack and into the designated register
// and then increments the stack pointer 2x
func pop(cpu *CPU) {
    value := cpu.mmu.read16(cpu.stackPointer)
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
    cpu.mmu.write16(cpu.stackPointer, value)
    cpu.programCounter++
}

// res - resets the n-th bit of the specified register
func res(cpu *CPU) {
    register := cpu.currentInstruction() & 0x7
    registerValue := cpu.GetRegisterValue(register)
    targetBit := (cpu.currentInstruction() >> 3) & 0x7
    cpu.SetRegister(register, registerValue & ^(0x1<<targetBit))
    // no flags are affected by this operation
    cpu.programCounter++
}

// ret - sets the programCounter to the value currently on the stack
func ret(cpu *CPU) {
    cpu.programCounter = cpu.mmu.read16(cpu.stackPointer)
    cpu.stackPointer += 2
}

// retcc - return if the given condition is true, otherwise don't
func retcc(cpu *CPU) {
    if cpu.CheckCondition() {
        ret(cpu)
    } else {
        cpu.programCounter++
    }
}

// reti - returns from interrupt and enables interrupts
func reti(cpu *CPU){
    ret(cpu)
    cpu.inte = true
}

// rl - Rotate N left through carry flag
func rl(cpu *CPU) {
    register := cpu.currentInstruction() & 0x7
    value := cpu.GetRegisterValue(register)

    bit7 := value >> 7
    value = (value << 1)
    if cpu.carry {
        value = value | 0x1
    }

    cpu.subtract = false
    cpu.halfCarry = false
    cpu.zero = value == 0
    if bit7 != 0 {
        cpu.carry = true
    } else {
        cpu.carry = false
    }

    cpu.programCounter++
    cpu.SetRegister(register, value)
}

// rla - Rotate A left through carry
func rla(cpu *CPU) {
    bit7 := cpu.ra >> 7
    cpu.ra = cpu.ra << 1
    if cpu.carry {
        cpu.ra = cpu.ra | 0x1
    }
    cpu.carry = false
    if bit7 != 0 {
        cpu.carry = true
    }
    cpu.subtract = false
    cpu.halfCarry = false
    // Gameboy CPU Manual specifies that the Zero flag is set if the result
    // is zero, but this causes Blargg's ROM to fail
    cpu.zero = false
    cpu.programCounter++
}

// rlc - Rotates the given register 1 left, old bit 7 to carry flag
func rlc(cpu *CPU) {
    register := cpu.currentInstruction() & 0x7
    value := cpu.GetRegisterValue(register)

    bit7 := value >> 7
    value = (value << 1) | bit7

    cpu.zero = (value == 0)
    cpu.halfCarry = false
    cpu.subtract = false
    if bit7 != 0 {
        cpu.carry = true
    } else {
        cpu.carry = false
    }
    cpu.SetRegister(register, value)
    cpu.programCounter++
}

// rlca - Rotate A left, Old bit 7 to carry flag
func rlca(cpu *CPU) {
    bit7 := cpu.ra >> 7
    cpu.ra = (cpu.ra << 1) | bit7 // Rotate bit 7 to bit 0
    // Gameboy CPU Manual specifies that the Zero flag is set if the result
    // is zero, but this causes Blargg's ROM to fail
    cpu.zero = false
    cpu.halfCarry = false
    cpu.subtract = false
    if bit7 == 0 {
        cpu.carry = false
    } else {
        cpu.carry = true
    }
    cpu.programCounter++
}

// rra - rotate the accumulator through the carry flag
// the carry flag contents are copied to bit 7
// this is the same instruction as CB 1F apparently
func rra(cpu *CPU) {
    oldCarry := cpu.carry
    if cpu.ra&0x1 == 0x1 {
        cpu.carry = true
    } else {
        cpu.carry = false
    }
    cpu.ra = cpu.ra >> 1
    if oldCarry {
        cpu.ra = cpu.ra | 0x80
    }
    // Gameboy CPU Manual specifies that the Zero flag is set if the result
    // is zero, but this causes Blargg's ROM to fail
    cpu.zero = false
    cpu.subtract = false
    cpu.halfCarry = false
    cpu.programCounter++
}

// rrc - Rotate n right, old bit 0 to carry flag
func rrc(cpu *CPU) {
    register := cpu.currentInstruction() & 0x7
    value := cpu.GetRegisterValue(register)
    bit0 := value & 0x1
    value = (value >> 1) | (bit0 << 7)

    cpu.halfCarry = false
    cpu.subtract = false
    cpu.zero = value == 0
    if bit0 == 0 {
        cpu.carry = false
    } else {
        cpu.carry = true
    }

    cpu.SetRegister(register, value)
    cpu.programCounter++
}

// rrca - rotate A right and send the old bit 0 to carry
func rrca(cpu *CPU) {
    bit0 := cpu.ra & 0x1
    cpu.ra = (cpu.ra >> 1) | (bit0 << 7)
    if bit0 == 0 {
        cpu.carry = false
    } else {
        cpu.carry = true
    }
    cpu.subtract = false
    cpu.halfCarry = false
    // Gameboy CPU Manual specifies that the Zero flag is set if the result
    // is zero, but this causes Blargg's ROM to fail
    cpu.zero = false
    cpu.programCounter++
}

// rrn - rotate the given register right through the carry flag
// the carry flag contents are copied to bit 7
func rrn(cpu *CPU) {
    register := cpu.currentInstruction() & 0x7
    value := cpu.GetRegisterValue(register)
    oldCarry := cpu.carry
    if value&0x1 == 0x1 {
        cpu.carry = true
    } else {
        cpu.carry = false
    }
    value = value >> 1
    if oldCarry { // previously set to 1
        value = value | 0x80 // set the MSB to 1
    }
    cpu.SetRegister(register, value)
    cpu.zero = value == 0x0
    cpu.subtract = false
    cpu.halfCarry = false
    cpu.programCounter++
}

// rst - push address on stack and then jump to address embeded in instruction
func rst(cpu *CPU){
    cpu.stackPointer -=2
    cpu.mmu.write16(cpu.stackPointer, cpu.programCounter+1)
    
    address := (cpu.currentInstruction() >> 3) & 0x7
    cpu.programCounter = uint16(address << 3)
}

// sbc - Subtract the given register's value from A with the carry bit
func sbc(cpu *CPU) {
    value := cpu.GetRegisterValue(cpu.currentInstruction() & 0x7)
    if cpu.carry {
        cpu.ra = cpu.Sub(cpu.ra, value, 1)
    } else {
        cpu.ra = cpu.Sub(cpu.ra, value, 0)
    }
    cpu.programCounter++
}

// sbcd8 - Subtract the immediate value AND the carry bit from A
func sbcd8(cpu *CPU) {
    value := cpu.immediate8()
    if cpu.carry {
        cpu.ra = cpu.Sub(cpu.ra, value, 1)
    } else {
        cpu.ra = cpu.Sub(cpu.ra, value, 0)
    }
    cpu.programCounter += 2
}

// scf - Set the carry flag
func scf(cpu *CPU) {
    cpu.carry = true
    //cpu.zero - not affected
    cpu.subtract = false
    cpu.halfCarry = false
    cpu.programCounter++
}

// sbi - Subtracts the immediate value from A and then stores it into A
func sbi(cpu *CPU) {
    cpu.ra = cpu.Sub(cpu.ra, cpu.immediate8(), 0)
    cpu.programCounter += 2
}

// set - sets the n-th bit of the specified register
func set(cpu *CPU) {
    register := cpu.currentInstruction() & 0x7
    registerValue := cpu.GetRegisterValue(register)
    targetBit := (cpu.currentInstruction() >> 3) & 0x7
    cpu.SetRegister(register, registerValue|(0x1<<targetBit))
    // no flags are affected by this operation
    cpu.programCounter++
}

// sla - Shift N left into carry, LSB of n set to 0
func sla(cpu *CPU) {
    register := cpu.currentInstruction() & 0x7
    value := cpu.GetRegisterValue(register)
    bit7 := value >> 7
    value = value << 1
    if bit7 != 0 {
        cpu.carry = true
    } else {
        cpu.carry = false
    }
    cpu.zero = value == 0
    cpu.halfCarry = false
    cpu.subtract = false
    cpu.SetRegister(register, value)
    cpu.programCounter++
}

// sra - Shift n right into carry, MSB does not change
func sra(cpu *CPU) {
    register := cpu.currentInstruction() & 0x7
    value := cpu.GetRegisterValue(register)
    bit0 := value & 0x1
    bit7 := value & 0x80
    value = (value >> 1) | bit7

    cpu.zero = value == 0
    cpu.halfCarry = false
    cpu.subtract = false
    if bit0 != 0 {
        cpu.carry = true
    } else {
        cpu.carry = false
    }
    cpu.SetRegister(register, value)
    cpu.programCounter++
}

// srl - shift the given register 1 bit to the right. the least significant
// bit gets shifted to the carry bit and the most significant bit is set to 0
func srl(cpu *CPU) {
    register := cpu.currentInstruction() & 0x7
    value := cpu.GetRegisterValue(register)
    if value&0x1 == 0x1 {
        cpu.carry = true
    } else {
        cpu.carry = false
    }
    value = value >> 1
    cpu.zero = value == 0x0
    cpu.subtract = false
    cpu.halfCarry = false
    cpu.SetRegister(register, value)
    cpu.programCounter++
}

// sub - Performs A - given register
func sub(cpu *CPU) {
    value := cpu.GetRegisterValue(cpu.currentInstruction() & 0x7)
    cpu.ra = cpu.Sub(cpu.ra, value, 0)
    cpu.programCounter++
}

// swap - Swaps the upper & lower nibbles of the given register
func swap(cpu *CPU) {
    register := cpu.currentInstruction() & 0x7
    value := cpu.GetRegisterValue(register)
    lower := value & 0xF
    value = (value >> 4) | (lower << 4)
    cpu.SetRegister(register, value)
    cpu.zero = value == 0x0
    cpu.subtract = false
    cpu.halfCarry = false
    cpu.carry = false
    cpu.programCounter++
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

// xord8 - Exclusive OR of the immediate value with the accumulator
func xord8(cpu *CPU) {
    cpu.ra = cpu.ra ^ cpu.immediate8()
    cpu.zero = (cpu.ra == 0)
    cpu.halfCarry = false
    cpu.subtract = false
    cpu.carry = false
    cpu.programCounter += 2
}

func unimplemented(cpu *CPU) {
    errStr := fmt.Sprintf("Instruction [%X] is not yet implemented", cpu.currentInstruction())
    panic(errStr)
}

func unimplementedExtended(cpu *CPU) {
    errStr := fmt.Sprintf("Extended Instruction [CB %X] is not yet implemented", cpu.currentInstruction())
    panic(errStr)
}

func (cpu * CPU) interrupt(interrupt uint8, address uint16){
    // Wake up the CPU if the CPU is in halted mode
    if cpu.halted{ // verbose
        cpu.halted = false
    }

    cpu.inte = false; // Disable the interrupt flag. Have to call RETI or EI to re-enable
    cpu.stackPointer -= 2
    cpu.mmu.write16(cpu.stackPointer, cpu.programCounter)
    cpu.programCounter = address
    newIF := cpu.mmu.getIF() & ^interrupt
    cpu.mmu.setIF(newIF)
}

// checkForInterrupts
// Check to see if an interrupt has occured (IF set from any source)
func (cpu * CPU) checkForInterrupts(){
    interruptFlag := cpu.mmu.getIF()
    interruptEnabledFlag := cpu.mmu.getIE() // Get enabled interrupts

    if interruptFlag == 0x0 {
        // No interrupts set so nothing to do here
        return 
    }

    if(cpu.inte){
        // Check to see if the interrupt handler associated with the interrupt is enabled.
        // If it is, go run the interrupt service routine
        if ((interruptFlag & 0x1) > 0) && ((interruptEnabledFlag & 0x1) > 0){ // V-BLANK
            cpu.interrupt(0x1, 0x40)           
        } else if ((interruptFlag & 2) > 0) && ((interruptEnabledFlag & 0x2) > 0){ // LCDC
            cpu.interrupt(0x2, 0x48)
        } else if ((interruptFlag & 4) > 0) && ((interruptEnabledFlag & 0x4) > 0) { // Timer overflow
            cpu.interrupt(0x4, 0x50)
        } else if ((interruptFlag & 8) > 0) && ((interruptEnabledFlag & 0x8) > 0) { // Serial transfer
            cpu.interrupt(0x8,0x58)
        } else if ((interruptFlag & 16) > 0) && ((interruptEnabledFlag & 0x16) >0) { // Hi-Lo of P10-P13 (button input)
            cpu.interrupt(0x16, 0x60)
        }
    } else {
        // Special code for handling HALT that was called when interrupts are not enabled
        // The interrupt will wake the CPU, but it will not be handled OR cleared
        if cpu.halted {
            if (interruptFlag & interruptEnabledFlag) != 0 {
                cpu.halted = false
                cpu.programCounter++
            }
        }
    }
}

// prettyDebugOutputAboutCurrentInstruction - Does what it says on the tin
func prettyDebugOutputAboutCurrentInstruction(cpu * CPU) {
    instruction := cpu.currentInstruction()

    instructionInfo := Instruction{}
    if instruction != 0xCB {
        instructionInfo = cpu.mainInstructions[instruction]
    } else {
        instruction := cpu.nextInstruction()
        instructionInfo = cpu.extendedInstructions[instruction]
    }

    debugPrint(cpu, instructionInfo.name, instructionInfo.dataSize)
}

func (cpu *CPU) step() int {
    prettyDebugOutputAboutCurrentInstruction(cpu)
    cyclesThisStep := 0
    if !cpu.halted{
        instruction := cpu.currentInstruction()
        instructionInfo := Instruction{}

        if instruction != 0xCB {
            instructionInfo = cpu.mainInstructions[instruction]
        } else {
            cpu.programCounter++
            instruction := cpu.currentInstruction()
            instructionInfo = cpu.extendedInstructions[instruction]
        }

        instructionInfo.function(cpu) // Execute the instruction
        cpu.instructionsExecuted++
        cyclesThisStep = instructionInfo.cycles
        cpu.timer.update(cyclesThisStep) // Update the timers (which may trigger interrupts)
    } else {
        // Special code to handle what to do if the CPU is halted
        // During a halt, the CPU is executing 4 cycles every update
        instructionInfo := cpu.mainInstructions[cpu.currentInstruction()]
        cyclesThisStep = instructionInfo.cycles
        cpu.timer.update(cyclesThisStep) 
    }

    return cyclesThisStep
}
