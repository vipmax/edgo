package process

import (
	"bufio"
	"context"
	. "fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"syscall"
	"testing"
	"time"
)

func TestProcess(t *testing.T) {

	process := NewProcess("/Users/max/opt/anaconda3/bin/python", "atest.py")
	process.Cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

	process.Start()

	go func() {
		for line := range process.Out {
			Println("Output:", line)
		}
	}()

	//Println("Process started with PID:", process.Cmd.Process.Pid)

	// Kill the process after 5 seconds
	time.Sleep(3 * time.Second)
	process.Stop()
	//Println("Process killed with PID:", process.Cmd.Process.Pid)

	time.Sleep(30 * time.Second)

	Println("Child process finished.")
}

func TestProcessKill(t *testing.T) {
	cmd := exec.Command("sleep", "100", "&")
	err := cmd.Start()
	if err != nil { Println(err) }

	Println("Process started with PID:", cmd.Process.Pid)

	// Kill the process after 5 seconds
	time.Sleep(5 * time.Second)

	err = cmd.Process.Kill()
	if err != nil { Println(err) }
	cmd.Process.Release()

	Println("Process killed with PID:", cmd.Process.Pid)
	time.Sleep(30 * time.Second)
}


func TestKill2(t *testing.T) {
	cmd := exec.Command("sleep", "100")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Start()

	Printf("Parent PID: %d\n", cmd.Process.Pid)

	time.Sleep(3 * time.Second)
	go func( ) {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os. Interrupt)
		<-sig
		signal.Reset()
	}()
	time.Sleep(30 * time.Second)
}

func TestKill3(t *testing.T) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	go run(ctx)

	time.Sleep(3 * time.Second)

	stop()

	time.Sleep(30 * time.Second)
}

func run(ctx context.Context) {
	cmd := exec.CommandContext(ctx, "python3", "atest.py")
	//cmd := exec.CommandContext(ctx, "go", "run", "cmd/test/main.go")
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

	stdout, _ := cmd.StdoutPipe()

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			Println(line)
		}

		Println("done")

	}()

	cmd.Start()
}


// Regular expression to match ANSI escape code for color (e.g., \e[1;31m)
var ansiColorPattern = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"


// Function to extract ANSI escape code colors from a string
func extractANSIColors(input string) string {
	regex := regexp.MustCompile(ansiColorPattern)
	return regex.ReplaceAllString(input, "")

	//matches := regex.FindAllStringSubmatch(input, -1)
	//return matches
}


func TestExtractANSIColors(t *testing.T) {
	testCases := []struct {
		input          string
		expectedColors int
	}{
		{
			input:          `This is \e[1;31ma red\e[0m text and \e[1;34mblue\e[0m text.`,
			expectedColors: 2,
		},
		{
			input:          "No ANSI escape codes here.",
			expectedColors: 0,
		},
		{
			input:          `This has a \e[1;32mgreen\e[0m color code and \e[0;35mpurple\e[0m code.`,
			expectedColors: 2,
		},
		{
			input:          `Multiple color codes in a single string: \e[1;33myellow\e[0m, \e[1;36mcyan\e[0m, \e[1;35mmagenta\e[0m.`,
			expectedColors: 3,
		},
		{
			input:          `Incomplete color code: \e[1;31mred\e[0.`,
			expectedColors: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			detectedColorCodes := extractANSIColors(testCase.input)

			if len(detectedColorCodes) != testCase.expectedColors {
				t.Errorf("Expected %d ANSI escape code colors, but got %d", testCase.expectedColors, len(detectedColorCodes))
			}
		})
	}
}
