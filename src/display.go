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

// (0,0)               (0,255)
//      +----------------+
//      |    (scrX,Scry) |
//      |--+      +------|
//      |  |      |      |
//      |--+      +------|
//      |                |
//      +----------------+
//
func (display * Display) drawBackgroundLine(ly uint8){
    tileY := ly + display.cpu.mmu.scrollY() // This will overflow as needed!
   
    for lcdX := uint8(0); lcdX < LCDWIDTH; lcdX++ {
        tileX := lcdX + display.cpu.mmu.scrollX() // This will overflow as needed
        color := display.cpu.mmu.backgroundPixelAt(tileX,tileY)
        pixel := int(ly)*int(LCDWIDTH) + int(lcdX)

        display.internalImage.Pix[4*pixel] = uint8((color >> 24) & 0xFF)
        display.internalImage.Pix[4*pixel+1] = uint8((color >> 16) & 0xFF)
        display.internalImage.Pix[4*pixel+2] = uint8((color >> 8) & 0xFF)
        display.internalImage.Pix[4*pixel+3] = uint8(color & 0xFF)
    }
}

func (display * Display) renderLine(ly uint8){
    display.drawBackgroundLine(ly)
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

//   0  80           252                  456 (CPU Clock)
//   +---+-----------+---------------------+  0
//   |   |           |                     |
//   |OAM|   Pixel   |      H-Blank        |
//   |   |  Transfer |                     |
//   |   |           |                     |
//   +---+-----------+---------------------+  144
//   |            V-Blank Time             |
//   +-------------------------------------+  154 (LCD Lines)
//
func (display * Display) calculateAndSetSTAT() {
    if display.cpu.mmu.getLY() >= 144 {
        display.cpu.mmu.setSTATMode(0x1) // In VBLANK
    } else {
        if display.scanlineCounter <= 80 {
            display.cpu.mmu.setSTATMode(0x2) // OAM search
        } else if display.scanlineCounter > 80 && display.scanlineCounter <= 252 {
            display.cpu.mmu.setSTATMode(0x3) // Pixel transfer state
        } else {
            display.cpu.mmu.setSTATMode(0x0) // H-Blank
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
    display.calculateAndSetSTAT()

    if display.scanlineCounter < 456 { // No new scanline just yet
        return
    }

    if !display.cpu.mmu.showDisplay(){ // LCDC bit 7 is disabled
        return
    }

    ly := display.cpu.mmu.getLY()
    if ly < 144 { // Can only render the first 144 rows - the rest are never rendered
        display.renderLine(ly)
    }

    //fmt.Println(y)
    //if y == 0 { // For testing purposes, render everything once per frame
    //   display.renderScreen()
    //}

    // Scanline ended here!
    display.scanlineCounter -= 456 // Save the extra cycles
    display.cpu.mmu.incrementLY()
}

func newDisplay(cpu *CPU) *Display {
    fmt.Println("Initializing Display")
    display := new(Display)
    display.cpu = cpu
    display.scanlineCounter = 0
    display.internalImage = image.NewRGBA(image.Rect(0, 0, int(LCDWIDTH), int(LCDHEIGHT) ))
    return display
}