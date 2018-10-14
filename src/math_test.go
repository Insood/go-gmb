package main

import "testing"

// MathTest : A structure which has input & output values for a math test
type Operation int

const (
    subtraction Operation = 1
    addition    Operation = 2
)

type MathTest struct {
    op        Operation // Type of math we're doing
    a         uint8     // In
    b         uint8     // In
    result    uint8     // Out
    zero      bool      // Out: 1 if result is 0x0
    carry     bool      // Out: 1 if carry out of the MSB
    halfCarry bool      // Out: 1 if carry out of MSB nibble bit
    subtract  bool      // Out: 1 if a subtraction was performed
}

//             operation,   a,   b  result, zero, carry, half, subtract
var subTests = []MathTest{
    MathTest{addition, 0x44, 0x11, 0x55, false, false, false, false},
    MathTest{addition, 0x23, 0x33, 0x56, false, false, false, false},
    //MathTest{addition, 0x2E, 0x74, 0xA2, false, false, false, true, true},
    //MathTest{addition, 0xA7, 0x59, 0x00, true, true, true, true, false},
    //MathTest{subtraction, 0x11, 0x11, 0x0, true, false, true, true, false},
    //MathTest{subtraction, 0xF5, 0xF5, 0x0, true, false, true, true, false},
    //MathTest{addition, 12, 0xF1, 0xFD, false, false, false, false, true}, // 0xF1 = -15, 0xFD is -3
    //MathTest{subtraction, 12, 15, 0xFD, false, true, false, false, true}, // 0xFD is -3
    //MathTest{subtraction, 197, 98, 99, false, false, true, true, false},
}

// TestMath - run a series of tests on the ALU
func TestMath(t *testing.T) {
    for _, test := range subTests {
        cpu := newCPU()
        var result uint8
        if test.op == subtraction {
            result = cpu.Sub(test.a, test.b, 0)
            t.Logf("%d - %d = %d\n", test.a, test.b, result)
        } else if test.op == addition {
            result = cpu.Add(test.a, test.b, 0)
            t.Logf("%d + %d = %d\n", test.a, test.b, result)
        }
        if result != test.result {
            t.Errorf("Result is incorrect. Expected: %X, Got %X", test.result, result)
        }
        if test.zero != cpu.zero {
            t.Errorf("Zero bit is incorrect. Expected %t, Got %t", test.zero, cpu.zero)
        }
        if test.carry != cpu.carry {
            t.Errorf("Carry bit is incorrect. Expected %t, Got %t", test.carry, cpu.carry)
        }
        if test.halfCarry != cpu.halfCarry {
            t.Errorf("halfCarry bit is incorrect. Exepected %t, Got %t", test.halfCarry, cpu.halfCarry)
        }
        if test.subtract != cpu.subtract {
            t.Errorf("Subtract bit is incorrect. Expected %t, Got %t", test.subtract, cpu.subtract)
        }
    }
}
