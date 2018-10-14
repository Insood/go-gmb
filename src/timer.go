package main

/*
import (
    "fmt"
)*/

// Timer - an interface for the MMU which handles updating timer-related registers
type Timer struct {
    mmu * MMU
    cyclesSinceLastTIMAUpdate int
    cyclesSinceLastDIVUpdate int
    cpuCycles uint64 // CPU timer operation undefined if left operational for more than 139,365 years due to 64-bit limitations
}

// cyclesPerTIMAUpdate - Get the update frequency from the TAC register
// A gameboy has a clock speed of 4,194,304 clock cycles per second.
// There are 4 timer settings: 4096, 16384, 65536, 262144 Hz
// Which translates to an interrupt every: 1024, 256, 64, and 16 cycles
func (timer * Timer) cyclesPerTIMAUpdate() int {
    tac  := timer.mmu.getTAC() // Timer Control Register
    frequencyTable := [4]int{1024, 16, 64, 256} // 0x0: 4096, 0x1: 2622144, 0x2: 65536, 0x3: 16384
    timerEnabled := ((tac >> 2) & 0x1 > 0)
    if( !timerEnabled){
        return 0;
    }
    return frequencyTable[tac & 0x3]
}

// updateTimers - updates the internal timers per specification
// Takes the number of cycles that were performed
func (timer * Timer) update(cyclesPerformed int){
    timer.cpuCycles += uint64(cyclesPerformed)

    // DIV (Divider Register) is updated at 16348Hz (256 cycles)
    // even if the main timer is disabled
    timer.cyclesSinceLastDIVUpdate += cyclesPerformed 
    if timer.cyclesSinceLastDIVUpdate > 256 {
        timer.mmu.incrementDIV()
        timer.cyclesSinceLastDIVUpdate -= 256
    }

    cyclesPerUpdate := timer.cyclesPerTIMAUpdate()
    if(cyclesPerUpdate == 0){ // Timer is disabled
        return 
    }

    timer.cyclesSinceLastTIMAUpdate += cyclesPerformed
     // Possible bug if the TAC frequency is changed down (ie: from 262144 to 4096Hz)
     // then the above counter may have enough "saved up" cycles to increment TIMA multiple
     // times.
    if timer.cyclesSinceLastTIMAUpdate > cyclesPerUpdate {
        timer.cyclesSinceLastTIMAUpdate -= cyclesPerUpdate
        tima := timer.mmu.getTIMA() // Timer Counter Register
        tima++

        // Another possible bug, whenever TIMA overflows, TMA is not set for another 4 clock cycles
        // see: http://gbdev.gg8.se/wiki/articles/Timer_Obscure_Behaviour
        // Also, IF is not set until the next clock either, but let's fake it for now
        if( tima == 0) { // overflow in TIMA occured
            timer.mmu.setTIMA(timer.mmu.getTMA())
            IF := timer.mmu.getIF()
            timer.mmu.setIF(IF | 0x4) // Enable bit 2 of the interrupt enable register
            //fmt.Println("Triggering IF")
        } else {
            timer.mmu.setTIMA(tima)
        }
    }
}

func createTimer(mmu *MMU) * Timer {
    timer := new(Timer)
    timer.mmu = mmu
    return timer
}