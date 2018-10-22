package main

import "testing"

func TestInstructionCyclesConditionNotTaken(t *testing.T){
    cpu := newCPU()
    // from Blargg's instr_timing.s file
    timings := [256]int{1,3,2,2,1,1,2,1,5,2,2,2,1,1,2,1,
                        0,3,2,2,1,1,2,1,3,2,2,2,1,1,2,1,
                        2,3,2,2,1,1,2,1,2,2,2,2,1,1,2,1,
                        2,3,2,2,3,3,3,1,2,2,2,2,1,1,2,1,
                        1,1,1,1,1,1,2,1,1,1,1,1,1,1,2,1,
                        1,1,1,1,1,1,2,1,1,1,1,1,1,1,2,1,
                        1,1,1,1,1,1,2,1,1,1,1,1,1,1,2,1,
                        2,2,2,2,2,2,0,2,1,1,1,1,1,1,2,1,
                        1,1,1,1,1,1,2,1,1,1,1,1,1,1,2,1,
                        1,1,1,1,1,1,2,1,1,1,1,1,1,1,2,1,
                        1,1,1,1,1,1,2,1,1,1,1,1,1,1,2,1,
                        1,1,1,1,1,1,2,1,1,1,1,1,1,1,2,1,
                        2,3,3,4,3,4,2,4,2,4,3,0,3,6,2,4,
                        2,3,3,0,3,4,2,4,2,4,3,0,3,0,2,4,
                        3,3,2,0,0,4,2,4,4,1,4,0,0,0,2,4,
                        3,3,2,1,0,4,2,4,3,2,4,1,0,0,2,4}
    for i := 0 ; i < 255; i++ {
        instruction := cpu.mainInstructions[i]
        if instruction.cyclesWhenBranchNotTaken != timings[i]*4 {
            t.Errorf("Main instruction %02X does not have the correct timing defined. Should be: %d, is: %d",
                     i, timings[i]*4, instruction.cyclesWhenBranchNotTaken)
        }
    }
}

func TestInstructionCyclesConditionTaken(t *testing.T){
    cpu := newCPU()
    // from Blargg's instr_timing.s file
    timings := [256]int{1,3,2,2,1,1,2,1,5,2,2,2,1,1,2,1,
                        0,3,2,2,1,1,2,1,3,2,2,2,1,1,2,1,
                        3,3,2,2,1,1,2,1,3,2,2,2,1,1,2,1,
                        3,3,2,2,3,3,3,1,3,2,2,2,1,1,2,1,
                        1,1,1,1,1,1,2,1,1,1,1,1,1,1,2,1,
                        1,1,1,1,1,1,2,1,1,1,1,1,1,1,2,1,
                        1,1,1,1,1,1,2,1,1,1,1,1,1,1,2,1,
                        2,2,2,2,2,2,0,2,1,1,1,1,1,1,2,1,
                        1,1,1,1,1,1,2,1,1,1,1,1,1,1,2,1,
                        1,1,1,1,1,1,2,1,1,1,1,1,1,1,2,1,
                        1,1,1,1,1,1,2,1,1,1,1,1,1,1,2,1,
                        1,1,1,1,1,1,2,1,1,1,1,1,1,1,2,1,
                        5,3,4,4,6,4,2,4,5,4,4,0,6,6,2,4,
                        5,3,4,0,6,4,2,4,5,4,4,0,6,0,2,4,
                        3,3,2,0,0,4,2,4,4,1,4,0,0,0,2,4,
                        3,3,2,1,0,4,2,4,3,2,4,1,0,0,2,4}
    for i := 0 ; i < 255; i++ {
        instruction := cpu.mainInstructions[i]
        if instruction.cycles != timings[i]*4 {
            t.Errorf("Main instruction %02X does not have the correct timing defined. Should be: %d, is: %d",
                     i, timings[i]*4, instruction.cycles)
        }
    }
}


func TestExtendedInstructionCycles(t *testing.T){
    cpu := newCPU()
    // from Blargg's instr_timing.s file
    timings := [256]int{2,2,2,2,2,2,4,2,2,2,2,2,2,2,4,2,
                        2,2,2,2,2,2,4,2,2,2,2,2,2,2,4,2,
                        2,2,2,2,2,2,4,2,2,2,2,2,2,2,4,2,
                        2,2,2,2,2,2,4,2,2,2,2,2,2,2,4,2,
                        2,2,2,2,2,2,3,2,2,2,2,2,2,2,3,2,
                        2,2,2,2,2,2,3,2,2,2,2,2,2,2,3,2,
                        2,2,2,2,2,2,3,2,2,2,2,2,2,2,3,2,
                        2,2,2,2,2,2,3,2,2,2,2,2,2,2,3,2,
                        2,2,2,2,2,2,4,2,2,2,2,2,2,2,4,2,
                        2,2,2,2,2,2,4,2,2,2,2,2,2,2,4,2,
                        2,2,2,2,2,2,4,2,2,2,2,2,2,2,4,2,
                        2,2,2,2,2,2,4,2,2,2,2,2,2,2,4,2,
                        2,2,2,2,2,2,4,2,2,2,2,2,2,2,4,2,
                        2,2,2,2,2,2,4,2,2,2,2,2,2,2,4,2,
                        2,2,2,2,2,2,4,2,2,2,2,2,2,2,4,2,
                        2,2,2,2,2,2,4,2,2,2,2,2,2,2,4,2}
    for i := 0 ; i < 255; i++ {
        instruction := cpu.extendedInstructions[i]
        if instruction.cycles != timings[i]*4 {
            t.Errorf("Extended instruction %02X does not have the correct timing defined. Should be: %d, is: %d",
                     i, timings[i]*4, instruction.cycles)
        }
    }
}