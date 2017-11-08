# Game Boy Emulator

Game Boy (LR35902) emulator implemented in golang. Based on my [8080](https://github.com/Insood/8080) emulator.

Dependencies:
1) Ebiten 2D library (https://github.com/hajimehoshi/ebiten)

Built in GO with lots of help from the following resources:
1) #ebiten on gopher.slack.com
2) Instructions in tabular format: http://www.pastraiser.com/cpu/gameboy/gameboy_opcodes.html
3) Memory layout: http://gbdev.gg8.se/wiki/articles/The_Cartridge_Header
4) Instruction encoding: http://www.classiccmp.org/dunfield/r/8080.txt

Known Issues:
1) Doesn't do anything yet