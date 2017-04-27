package main

import "os/exec"
import "log"
import "os"

import "encoding/json"
import "fmt"
import "io"
import "strings"

// Executable to be executed
type Executable struct {
	Type    string
	Version string
	Command []string
}

// The marker of exe meta data
var ExeMarker = []byte{0xda, 0xf0, 0x47, 0x5c}

// The max offset of the marker from the tail of the exe file
var ExeMarkerMaxOffset = 1024

// Whether to enable debug
var debugEnabled = false

var thisExeFilePath = ""

func init() {
	var err error
	var arg0 = os.Args[0]

	thisExeFilePath, err = exec.LookPath(arg0)
	if err != nil {
		thisExeFilePath = arg0
	}
}

func main() {
	args := os.Args

	args = enableDebugModeIfNeeded(args)

	exe := tryParseMetaFromOsExe(args)
	debugf("try parsed exe: %v", exe)

	if exe != nil {
		exe.exec(args[1:])
		return
	}

	if len(args) <= 1 {
		printHelp()
		return
	}

	if args[1] == "-o" {
		if len(args) < 4 {
			printHelp()
			return
		}

		generateExeTo(args[2], &Executable{
			Type:    "exe",
			Version: "1.0",
			Command: args[3:],
		})
		return
	}

	printHelp()
	return
}

func debugf(fmt string, v ...interface{}) {
	if debugEnabled {
		v = append([]interface{}{os.Getpid()}, v...)
		log.Printf("[%v] "+fmt, v...)
	}
}

func (exe *Executable) exec(extraArgs []string) *os.ProcessState {
	exeFile, err := exec.LookPath(exe.Command[0])
	if err != nil {
		exeFile = exe.Command[0]
	}

	exeArgs := append(exe.Command[1:], extraArgs...)
	cmd := exec.Command(exeFile, exeArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		log.Panic(err)
		return nil
	}

	// log.Print("Command executing...")

	err = cmd.Wait()

	// log.Printf("Command finished with error: %v", err)
	if err != nil {
		panic(err)
		return nil
	}

	return cmd.ProcessState
}

func tryParseMetaFromOsExe(args []string) *Executable {
	debugf("args: %v", args)

	file, err := os.OpenFile(thisExeFilePath, os.O_RDONLY, 0755)
	if err != nil {
		debugf("Failed to open exe file: %v", err)
		return nil
	}

	defer file.Close()

	file.Seek(int64(-ExeMarkerMaxOffset), os.SEEK_END)

	buf := make([]byte, ExeMarkerMaxOffset)
	readBytes, err := file.Read(buf)
	if err != nil {
		debugf("Failed to read exe file: %v", err)
		return nil
	}

	debugf("readBytes: %v", readBytes)

	marker := ExeMarker
	i := readBytes - 5
	for ; i > 0; i-- {
		if buf[i] == marker[0] && buf[i+1] == marker[1] && buf[i+2] == marker[2] && buf[i+3] == marker[3] {
			break
		}
	}

	meta := buf[i+4 : readBytes]

	debugf("Got meta[%v:%v]: %v", i, readBytes, string(meta))

	var exe Executable

	err = json.Unmarshal(meta, &exe)
	if err != nil {
		debugf("Failed to parse JSON: %v", err)
		return nil
	}

	if &exe == nil {
		debugf("What? Exe is nil!")
		return nil
	}

	return &exe
}

func printHelp() {
	fmt.Println("Usage: exe-delegate -o <delegate-exe-file> <original-exe-file-path> [...args]")
	os.Exit(1)
}

func generateExeTo(outputFile string, exe *Executable) {
	debugf("generateExeTo - Trying to serialize exe info: %v", exe)

	meta, err := json.Marshal(exe)
	if err != nil {
		panic(err)
	}

	exeFile1, err := os.OpenFile(thisExeFilePath, os.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}

	defer exeFile1.Close()

	if !strings.HasSuffix(strings.ToLower(outputFile), ".exe") {
		outputFile += ".exe"
	}

	outputFileObj, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}

	defer outputFileObj.Close()

	_, err = io.Copy(outputFileObj, exeFile1)
	if err != nil {
		panic(err)
	}

	_, err = outputFileObj.Write(ExeMarker)
	if err != nil {
		panic(err)
	}

	_, err = outputFileObj.Write(meta)
	if err != nil {
		panic(err)
	}

	outputFileObj.Sync()

	debugf("generateExeTo - all written.")
}

func enableDebugModeIfNeeded(args []string) []string {
	for i, n := 1, len(args); i < n; i++ {
		if args[i] == "--debug-exe-delegate" {
			debugEnabled = true
			return append(append([]string{}, args[0:i]...), args[i+1:]...)
		}
	}

	return args
}
