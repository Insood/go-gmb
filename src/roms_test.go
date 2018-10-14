package main

import ( 
    //"fmt"
    "bytes"
    "testing"
    "os/exec"
    "time"
    "io/ioutil"
    "strings"
)

func checkOutput(t * testing.T, romName string, output string){
    if !strings.Contains(output,"Passed") {
        t.Errorf("ROM %s did not get a passing result", romName)
    }
}

func runROM(t *testing.T, directory string, romName string){
    cmdName := "src.exe"
    fullROMName := directory + "/" + romName

    cmdArgs := []string{"-v=false",fullROMName}
    cmd := exec.Command(cmdName,cmdArgs...)

    // From https://medium.com/@vCabbage/go-timeout-commands-with-os-exec-commandcontext-ba0c861ed738
    var buf bytes.Buffer
    cmd.Stdout = &buf

    cmd.Start()

    // Use a channel to signal completion so we can use a select statement
    done := make(chan error)
    go func() { done <- cmd.Wait() }()

    // Start a timer
    timeout := time.After(500 * time.Millisecond)

    select {
    case <-timeout:
        cmd.Process.Kill()
        checkOutput(t,romName, buf.String())
    case <-done:
        checkOutput(t,romName, buf.String())
    }
}

// TestROMs - Automatically runs the test ROMs and turns them into test case results
func TestROMs(t *testing.T) {
    romDirectory := "../gb-test-roms/cpu_instrs/individual"
    files, err := ioutil.ReadDir(romDirectory)
    if err != nil {
        t.Errorf("Could not open directory %s",romDirectory)
    }
    for _, file := range files {
        runROM(t, romDirectory, file.Name())
    }
}
