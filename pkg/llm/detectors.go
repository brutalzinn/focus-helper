package llm

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

func AnalyzeHyperfocus(
	adapter LLMAdapter,
	promptTemplate string,
	numLevels int,
	history string,
	currentWindow string,
	usageDuration time.Duration,
) (int, error) {
	maxIndex := numLevels - 1
	prompt := fmt.Sprintf(promptTemplate, numLevels, maxIndex, maxIndex)
	context := fmt.Sprintf(
		"Current Time: %s\nUser History: %s\nCurrently Focused Window Title: '%s'\nContinuous Usage Duration: %s",
		time.Now().Format(time.RFC1123),
		history,
		currentWindow,
		usageDuration.Round(time.Second).String(),
	)
	fullPrompt := fmt.Sprintf("%s\n\n--- CURRENT SITUATION ---\n%s\n\nBased on this data, what is the appropriate alert index (from 0 to %d, or -1 for none)?", prompt, context, maxIndex)
	log.Println("Sending context to AI Analyst via adapter...")
	response, err := adapter.Generate(fullPrompt)
	if err != nil {
		return -1, fmt.Errorf("AI adapter failed to generate response: %w", err)
	}
	indexStr := strings.TrimSpace(response)
	indexStr = strings.Trim(indexStr, "`'\" .")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return -1, fmt.Errorf("AI returned a non-numeric response: '%s'", indexStr)
	}
	if index >= numLevels {
		return -1, fmt.Errorf("AI returned an out-of-bounds index: %d (max is %d)", index, maxIndex)
	}
	return index, nil
}
