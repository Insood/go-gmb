package main

// SCREENWIDTH - Horizontal resolution of the display (columns of pixels)
var SCREENWIDTH = 256

// SCREENHEIGHT - Vertical Resolution of the display (rows of pixels)
var SCREENHEIGHT = 256

// SCREENSCALE - How much to upscale the display to fit the monitor better
var SCREENSCALE = float64(2)

// CYCLESPERFRAME - How many cycles to run every frame
// Ebit renders at ~60fps while the GB renders at ~59.7
// Technically the emulator will be running 0.5% faster
var CYCLESPERFRAME = 70224

/*DisplayPixel::White => RGB8 {
	r: 0x9B,
	g: 0xBC,
	b: 0x0F,
},
DisplayPixel::LightGrey => RGB8 {
	r: 0x8B,
	g: 0xAC,
	b: 0x0F,
},
DisplayPixel::DarkGrey => RGB8 {
	r: 0x30,
	g: 0x62,
	b: 0x30,
},
DisplayPixel::Black => RGB8 {
	r: 0x0F,
	g: 0x38,
	b: 0x0F,
}*/