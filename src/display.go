package main

import "image"

// Display - represents the LCD of the game boy
type Display struct {
    cpu * CPU
    scanlineCounter int
    internalImage *image.RGBA
}

func (display * Display) renderLine(y uint8){

}

// updateDisplay - Updates the display by taking in the number of cycles that the last
// instruction took to render.
// A single scanline takes 456 cycles to 'render'.
// This emulator will draw to the canvas every after scanline so that the original GB-like
// rendering is preserved
func (display *Display) updateDisplay(cycles int){
    display.scanlineCounter += cycles
    if display.scanlineCounter < 456 { // No new scanline just yet
        return
    }

    y := display.cpu.mmu.getLY()
    if y < 144 { // Can only render the first 144 rows - the rest are never rendered
        display.renderLine(y)
    }

    // Scanline ended here!
    display.scanlineCounter -= 456 // Save the extra cycles
    display.cpu.mmu.incrementLY()
}

func newDisplay(cpu *CPU) *Display {
    display := new(Display)
    display.cpu = cpu
    display.scanlineCounter = 0
    display.internalImage = image.NewRGBA(image.Rect(0, 0, SCREENHEIGHT, SCREENWIDTH))
    return display
}