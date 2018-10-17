package main

import ( 
    "image"
    "fmt" 
)

// Display - represents the LCD of the game boy
type Display struct {
    cpu * CPU
    scanlineCounter int
    internalImage *image.RGBA
}

func (display * Display) renderLine(y uint8){

}

func (display * Display) readTile(tileNumber uint8) []uint8 {
    tileAddress := 0x8000 + (uint16(tileNumber) * 16) // 16 bytes per tile
    return display.cpu.mmu.cart.memory[tileAddress:tileAddress+16]
}

func (display * Display) drawTile(xTile int, yTile int, tileData []uint8){
    screenX := xTile*16
    screenY := yTile*16
    for y:= 0; y<8; y++{
        pxLow := tileData[y*2]
        pxHigh := tileData[y*2+1]

        for  x := uint8(0) ; x<8; x++ {
            if (((pxLow >> (7-x)) & 0x1) | ((pxHigh >> (7-x))& 0x1)) > 0 {
                pixel := (screenY+y)*256 + (screenX+int(x))
                //fmt.Printf("Pixel: (%d)",pixel)
                display.internalImage.Pix[4*pixel] = 0xFF
                display.internalImage.Pix[4*pixel+1] = 0xFF
                display.internalImage.Pix[4*pixel+2] = 0xFF
                display.internalImage.Pix[4*pixel+3] = 0xFF
            }
        }
    }
}

func (display * Display) renderScreen(){
    for y:= 0; y<32; y++{
        for x:=0; x<32; x++{
            address := uint16(0x9800 + y*32 + x)
            tileNumber := display.cpu.mmu.read8(address)
            tileData := display.readTile(tileNumber)
            display.drawTile(x,y,tileData)
        }
    }
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

    if y == 0 { // For testing purposes, render everything once per frame
        display.renderScreen()
    }

    // Scanline ended here!
    display.scanlineCounter -= 456 // Save the extra cycles
    display.cpu.mmu.incrementLY()
}

func newDisplay(cpu *CPU) *Display {
    fmt.Println("Initializing Display")
    display := new(Display)
    display.cpu = cpu
    display.scanlineCounter = 0
    display.internalImage = image.NewRGBA(image.Rect(0, 0, SCREENHEIGHT, SCREENWIDTH))
    return display
}