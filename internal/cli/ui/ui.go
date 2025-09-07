package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	Bold   = "\033[1m"
)

// colorize adds color to text if terminal supports it
func colorize(color, text string) string {
	if isTerminal() {
		return color + text + Reset
	}
	return text
}

// isTerminal checks if output is to a terminal
func isTerminal() bool {
	// Simple check - in production this would be more sophisticated
	return os.Getenv("TERM") != ""
}

// Header prints a prominent header
func Header(text string) {
	fmt.Println(colorize(Bold+Cyan, text))
}

// Success prints a success message
func Success(text string) {
	fmt.Println(colorize(Green, text))
}

// Error prints an error message
func Error(text string) {
	fmt.Println(colorize(Red, text))
}

// Warning prints a warning message
func Warning(text string) {
	fmt.Println(colorize(Yellow, text))
}

// Info prints an info message
func Info(text string) {
	fmt.Println(text)
}

// Debug prints a debug message (only if verbose)
func Debug(text string) {
	if os.Getenv("GOJANGO_DEBUG") == "1" {
		fmt.Println(colorize(Purple, "DEBUG: "+text))
	}
}

// Input prompts for user input with a default value
func Input(prompt, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}
	
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	
	return input
}

// Confirm prompts for yes/no confirmation
func Confirm(prompt string, defaultValue bool) bool {
	defaultStr := "y/N"
	if defaultValue {
		defaultStr = "Y/n"
	}
	
	fmt.Printf("%s [%s]: ", prompt, defaultStr)
	
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}
	
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return defaultValue
	}
	
	return input == "y" || input == "yes"
}

// Select prompts user to select from a list of options
func Select(prompt string, options []string) string {
	fmt.Println(prompt)
	for i, option := range options {
		fmt.Printf("  %d) %s\n", i+1, option)
	}
	
	for {
		fmt.Print("Select (1-" + strconv.Itoa(len(options)) + "): ")
		
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}
		
		input = strings.TrimSpace(input)
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(options) {
			Error("Invalid selection. Please try again.")
			continue
		}
		
		return options[choice-1]
	}
}

// MultiSelect allows selecting multiple options from a list
func MultiSelect(prompt string, options []string) []string {
	fmt.Println(prompt)
	fmt.Println("(Enter comma-separated numbers, or press Enter for none)")
	
	for i, option := range options {
		fmt.Printf("  %d) %s\n", i+1, option)
	}
	
	fmt.Print("Select: ")
	
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return []string{}
	}
	
	input = strings.TrimSpace(input)
	if input == "" {
		return []string{}
	}
	
	var selected []string
	parts := strings.Split(input, ",")
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		choice, err := strconv.Atoi(part)
		if err != nil || choice < 1 || choice > len(options) {
			continue
		}
		
		selected = append(selected, options[choice-1])
	}
	
	return selected
}

// ProgressBar displays a simple progress bar (placeholder for now)
func ProgressBar(current, total int, message string) {
	if total == 0 {
		return
	}
	
	percentage := (current * 100) / total
	fmt.Printf("\r%s... %d%% (%d/%d)", message, percentage, current, total)
	
	if current == total {
		fmt.Println()
	}
}