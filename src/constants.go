package main

// SCREENWIDTH - Horizontal resolution of the display (columns of pixels)
//var SCREENWIDTH = 256

// SCREENHEIGHT - Vertical Resolution of the display (rows of pixels)
//var SCREENHEIGHT = 256

// LCDWIDTH - The width of the display (in pixels)
var LCDWIDTH = uint8(160)

// LCDHEIGHT - The hight of the display (in pixels)
var LCDHEIGHT = uint8(144)


// SCREENSCALE - How much to upscale the display to fit the monitor better
var SCREENSCALE = float64(2)

// CYCLESPERFRAME - How many cycles to run every frame
// Ebit renders at ~60fps while the GB renders at ~59.7
// Technically the emulator will be running 0.5% faster
var CYCLESPERFRAME = 70224

// GameBoyColorMap - The 2-bit color palette to display on the screen
var GameBoyColorMap = []int { 0xFFFFFFFF, 0xB6B6B6FF, 0x676767FF, 0x000000FF}
// Lots of alternate palettes available here: https://lospec.com/palette-list/tag/gameboy
//var GameBoyColorMap = []int{ 0x9BBC0FFF, 0x8BAC0FFF, 0x306230FF, 0x0F380FFF} // DMG-like color