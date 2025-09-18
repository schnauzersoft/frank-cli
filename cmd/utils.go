/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// confirmAction prompts the user for confirmation
func confirmAction(action, scope string) bool {
	// Determine what to show in the prompt
	var promptText string
	if scope == "" {
		promptText = fmt.Sprintf("Do you want to %s all stacks? [y/N]", action)
	} else {
		promptText = fmt.Sprintf("Do you want to %s '%s'? [y/N]", action, scope)
	}

	fmt.Print(promptText + " ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	return response == "y" || response == "yes"
}
