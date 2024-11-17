package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/term"
)

type Message struct {
	Content string
	Author  string
	ID      int
}

var lock sync.Mutex

func clearScreen() error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func DrawTerminal(messages []Message) error {
	lock.Lock()
	defer lock.Unlock()

	width, _, err := term.GetSize(0)
	if err != nil {
		return fmt.Errorf("failed to get terminal size: %w", err)
	}

	if err := clearScreen(); err != nil {
		return fmt.Errorf("failed to clear screen: %w", err)
	}

	for _, msg := range messages {
		content := wordWrap(msg.Content, width-4)
		fmt.Printf("From %s: %s\n", msg.Author, content)
		fmt.Println(strings.Repeat("-", width-2))
	}

	fmt.Printf("> ")
	return nil
}

func wordWrap(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	currentLine := words[0]
	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)
	return strings.Join(lines, "\n")
}

func GetUserInput() string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}
