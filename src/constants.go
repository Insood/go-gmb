package main

// SCREENWIDTH - Horizontal resolution of the display (columns of pixels)
var SCREENWIDTH = 160

// SCREENHEIGHT - Vertical Resolution of the display (rows of pixels)
var SCREENHEIGHT = 144

// SCREENSCALE - How much to upscale the display to fit the monitor better
var SCREENSCALE = float64(2)

// CYCLESPERFRAME - How many cycles to run every frame
// Ebit renders at ~60fps while the GB renders at ~59.7
// Technically the emulator will be running 0.5% faster
var CYCLESPERFRAME = 70224