package main

import (
    "flag"
    "fmt"
    "os"
    "github.com/hajimehoshi/ebiten"
)

// DEBUGMODE - Whether or not the program is running in debug mode (ie: pretty print opcodes)
var DEBUGMODE = true

// ENABLEDISPLAY - Whether or not to render a display
var ENABLEDISPLAY = true

func debugPrintHeader(cpu *CPU) {
    if cpu.instructionsExecuted%20 == 0 {

        fmt.Printf("ADDR : %-27sB  C  D  E  H  L  A  ZNHC---- SP\n","instruction")
    }
}

// Outputs to stdout if DEBUGMODE is set
func debugPrintLn(str string) {
    if DEBUGMODE {
        fmt.Println(str)
    }
}

// debugPrint - will output a single line to the console regarding the current instruction
// Format:
// INST : PC <values> <instruction name> RB RC RD RE RH RL RA PSW SP
func debugPrint(cpu *CPU, name string, values int) {
    if !DEBUGMODE {
        return
    }

    debugPrintHeader(cpu)

    output := ""

    // Hard-wire an 0xCB before printing extended mode instructions
    cmd := ""
    if cpu.mmu.read8(cpu.programCounter-1) != 0xCB {
        cmd = fmt.Sprintf("%04X : %02X", cpu.programCounter, cpu.currentInstruction())
    } else {
        cmd = fmt.Sprintf("%04X : CB %02X", cpu.programCounter-1, cpu.currentInstruction())
    }

    for i := 1; i < values; i++ {
        cmd += fmt.Sprintf(" %02X", cpu.mmu.read8(cpu.programCounter+uint16(i)))
    }
    output += fmt.Sprintf("%-17s %-16s", cmd, name)

    //                      rb  rc   rd   re   rh   rl   ra   psw  SP
    output += fmt.Sprintf("%02X %02X %02X %02X %02X %02X %02X %08b %04X %02X %v\n",
        cpu.rb, cpu.rc, cpu.rd, cpu.re, cpu.rh, cpu.rl, cpu.ra, cpu.pswByte(), cpu.stackPointer, cpu.mmu.getTIMA(), cpu.timer.cpuCycles)

    fmt.Print(output)
}

func startup() string {
    args := os.Args[1:]
    if len(args) == 0 {
        fmt.Printf("%s <romname> - Runs the ROM <romname>", os.Args[0])
        os.Exit(0)
    }

    romName := args[len(args)-1]

    // Parse command line flags
    verboseFlag := flag.Bool("v", false, "Show every instruction being executed (slow)")
    displayFlag := flag.Bool("d", true, "Shows a display")
    flag.Parse()
    DEBUGMODE = *verboseFlag // Sadly - a global
    ENABLEDISPLAY =*displayFlag // Also another sad flag
    
    return romName
}

func generateTitle() string {
    return fmt.Sprintf("Go-GMB Emulator (%f) FPS",ebiten.CurrentFPS())
}

// debugMain - This is the loop that will run when the program starts with -d=false
// Display/sound are not enabled and the emulator runs as fast as possible
func debugMain(romName string){
    cpu := newCPU()
    cpu.mmu.cart = loadCart(romName)
    for {
        cpu.step()
        cpu.checkForInterrupts()
    }
}

// displayMain - This is the main emulator mode w/ a display & sound enabled
func displayMain(romName string){
    cpu := newCPU()
    cpu.mmu.cart = loadCart(romName)
    display := newDisplay(cpu)

    f := func(screen *ebiten.Image) error {
        cycleCounter := 0 // May cause up to ~28 extra cycles to be run during this frame render
        
        for cycleCounter <= CYCLESPERFRAME {
            cycles := cpu.step()
            display.updateDisplay(cycles) // This may trip interrupts so it goes before the interrupt dispatching function
            cpu.checkForInterrupts()
            cycleCounter += cycles
        }
        screen.ReplacePixels(display.internalImage.Pix)

        ebiten.SetWindowTitle(generateTitle())
        return nil
    }

    // Setup the main loop
    ebiten.SetRunnableInBackground(true)
    runErr := ebiten.Run(f, SCREENWIDTH, SCREENHEIGHT, SCREENSCALE, "Go-GMB Emulator")
    errStr := fmt.Sprintf("Exited run() with error: %s", runErr)
    fmt.Println(errStr)
}

func main() {
    romName := startup()

    if ENABLEDISPLAY {
        displayMain(romName)
    } else {
        debugMain(romName)
    }
}
