# Game Boy Emulator

Game Boy (LR35902) emulator implemented in golang. Based on my [8080](https://github.com/Insood/8080) emulator.

Dependencies:
1) Ebiten 2D library (https://github.com/hajimehoshi/ebiten)

Built in GO with lots of help from the following resources:
1) #gmb on emudev.slack.com
2) #ebiten on gopher.slack.com
3) Instructions in tabular format: http://www.pastraiser.com/cpu/gameboy/gameboy_opcodes.html
4) Memory layout: http://gbdev.gg8.se/wiki/articles/The_Cartridge_Header
5) Instruction encoding: http://www.classiccmp.org/dunfield/r/8080.txt
6) Test ROMs: https://github.com/retrio/gb-test-roms/
7) Opcode Summary: http://www.devrs.com/gb/files/opcodes.html


Blargg's cpu_instr test rom status:

  Test |  1  |  2  |  3  |  4  |  5  |  6  |  7  |  8  |  9  | 10  | 11 
   --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | ---
Status |  P  |  F  |  P  |  P  |  P  |  P  |  P  |  P  |  P  |  P  |  P

The unit tests have a common framework which require a bunch of instructions to be implemented ahead of time, but the list below is probably the order in which the unit tests can be attacked start to finish until they all pass.

Easy   - 3 (Basic stack pointer operations)
       - 6 (Basic 8-bit loads between registers)
       - 5 (Basic register operations)
       - 10 (Extended instructions; register bit setting)
       - 7 (Jump, call, return)
       - 8 (Push/Pop, 16-bit loads, high memory operations)
       - 11 (Extended instructions)
       - 4 (Immediate instructions: Load, Math, Bit-wise ops)
       - 9 (Math operations on registers; More extended instructions)
       - 01 (DAA)
Hard   - 02 (Interrupts)