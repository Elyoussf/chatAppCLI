package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

// findTerminal tries different terminal emulators and returns the first one found
func findTerminal() (string, []string, error) {
	if runtime.GOOS == "linux" {
		// List of common terminal emulators with their launch arguments
		terminals := []struct {
			name string
			args []string
		}{
			{"gnome-terminal", []string{"--"}},
			{"konsole", []string{"-e"}},
			{"xterm", []string{"-e"}},
			{"x-terminal-emulator", []string{"-e"}}, // Debian/Ubuntu default
			{"terminator", []string{"-x"}},
			{"urxvt", []string{"-e"}},
		}

		for _, term := range terminals {
			if path, err := exec.LookPath(term.name); err == nil {
				return path, term.args, nil
			}
		}
		return "", nil, fmt.Errorf("no supported terminal emulator found")
	}
	return "", nil, nil // Non-Linux systems don't need this
}

func runInNewTerminal(command string, args ...string) error {
	var terminalCmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		terminalPath, terminalArgs, err := findTerminal()
		if err != nil {
			return fmt.Errorf("failed to find terminal emulator: %v", err)
		}

		// Construct the full command
		fullArgs := append([]string{}, terminalArgs...) // Start with terminal's own args
		fullArgs = append(fullArgs, command)            // Add the command
		fullArgs = append(fullArgs, args...)            // Add command's args

		terminalCmd = exec.Command(terminalPath, fullArgs...)

	case "darwin":
		// For macOS, using Terminal.app
		cmdString := command
		for _, arg := range args {
			cmdString += " " + arg
		}
		terminalCmd = exec.Command("osascript", "-e",
			fmt.Sprintf(`tell app "Terminal" to do script "%s"`, cmdString))

	case "windows":
		// For Windows, create a slice to hold all arguments
		winArgs := []string{"/C", "start"}
		// Add the command and its arguments
		winArgs = append(winArgs, command)
		winArgs = append(winArgs, args...)
		terminalCmd = exec.Command("cmd.exe", winArgs...)

	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	// Set up environment if needed
	// terminalCmd.Env = append(os.Environ(), "DISPLAY=:0")

	// Start the command and return any error
	err := terminalCmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start terminal: %v", err)
	}
	return nil
}

func main() {
	// Launch server in first terminal
	fmt.Println("Launching server terminal...")
	err := runInNewTerminal("./Server/server.go")
	if err != nil {
		fmt.Printf("Error starting server terminal: %v\n", err)
		return
	}
	fmt.Println("Server terminal launched successfully")

	// Small delay to prevent potential race conditions
	//time.Sleep(time.Second)

	// Launch client in second terminal
	fmt.Println("Launching client terminal...")
	err = runInNewTerminal("go", "run", "./Client/client.go")
	if err != nil {
		fmt.Printf("Error starting client terminal: %v\n", err)
		return
	}
	fmt.Println("Client terminal launched successfully")
}
