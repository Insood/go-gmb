package main

import (
    "flag"
    "fmt"
    "os"
)

// DEBUGMODE - Whether or not the program is running in debug mode (ie: pretty print opcodes)
var DEBUGMODE = true

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

func main() {
    args := os.Args[1:]
    if len(args) == 0 {
        fmt.Printf("%s <program> - Runs the test program <program>", os.Args[0])
        return
    }

    romName := args[len(args)-1]

    // Parse command line flags
    verboseFlag := flag.Bool("v", true, "Show every instruction being executed (slow)")
    flag.Parse()

    DEBUGMODE = *verboseFlag
    cpu := newCPU()
    cpu.mmu.cart = loadCart(romName)

    for {
        cpu.step()
    }
}
