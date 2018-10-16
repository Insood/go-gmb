# Game Boy Emulator

Game Boy (LR35902) emulator implemented in golang. Based on my [8080](https://github.com/Insood/8080) emulator.

Dependencies:
1) Ebiten 2D library (https://github.com/hajimehoshi/ebiten)

Built in GO with lots of help from the following resources:
1) #gmb on emudev.slack.com
2) EmuDev discord
3) #ebiten on gopher.slack.com
4) Instructions in tabular format: http://www.pastraiser.com/cpu/gameboy/gameboy_opcodes.html
5) Memory layout: http://gbdev.gg8.se/wiki/articles/The_Cartridge_Header
6) Instruction encoding: http://www.classiccmp.org/dunfield/r/8080.txt
7) Test ROMs: https://github.com/retrio/gb-test-roms/
8) Opcode Summary: http://www.devrs.com/gb/files/opcodes.html
9) Information about HALT: https://github.com/AntonioND/giibiiadvance/tree/master/docs
10) http://www.codeslinger.co.uk/pages/projects/gameboy.html

Blargg's cpu_instr test rom status:

  Test |  1  |  2  |  3  |  4  |  5  |  6  |  7  |  8  |  9  | 10  | 11 
   --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | ---
Status |  P  |  P  |  P  |  P  |  P  |  P  |  P  |  P  |  P  |  P  |  P

The unit tests have a common framework which require a bunch of instructions to be implemented ahead of time, but the list below is probably the order in which the unit tests can be attacked start to finish until they all pass.

- 03 (Easiest) Basic stack pointer operations
- 06 Basic 8-bit loads between registers
- 05 Basic register operations
- 10 Extended instructions; register bit setting
- 07 Jump, call, return
- 08 Push/Pop, 16-bit loads, high memory operations
- 11 Extended instructions
- 04 Immediate instructions: Load, Math, Bit-wise ops. Half carry flag is a major PITA.
- 09 Math operations on registers; More extended instructions
- 01 DAA. You thought half carry was complicated?
- 02 (Hardest) Interrupts
