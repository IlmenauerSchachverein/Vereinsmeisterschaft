package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Converts pairings into a compact Markdown table
func convertRundeToMarkdown(runde string) string {
	lines := strings.Split(runde, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}
	if len(cleanLines) < 2 {
		return runde // Not enough data to create a table
	}

	// Compact headers for the table
	headers := []string{
		"Tisch", "TNr", "Teilnehmer", "Punkte", "-", "TNr", "Teilnehmer", "Punkte", "Ergebnis",
	}

	mdLines := []string{
		"| " + strings.Join(headers, " | ") + " |",
		"| " + strings.Repeat("--- | ", len(headers)),
	}

	// Process data rows
	for _, line := range cleanLines[1:] {
		columns := strings.Split(line, "\t")

		// Assemble the result (columns 10 and 12)
		ergebnis := columns[10] + " - " + columns[12]

		// Create a compact row
		row := []string{
			columns[0], // Table
			columns[1], // White TNr
			columns[2], // White Player
			columns[4], // White Points
			"-",        // Separator
			columns[6], // Black TNr
			columns[7], // Black Player
			columns[9], // Black Points
			ergebnis,   // Result
		}

		mdLines = append(mdLines, "| "+strings.Join(row, " | ")+" |")
	}

	return strings.Join(mdLines, "\n")
}

// Processes all .md files in the content folder
func processFiles(root string) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Fehler beim Zugriff auf Datei:", err)
			return nil
		}

		if !info.IsDir() && filepath.Ext(path) == ".md" {
			processFile(path)
		}
		return nil
	})
}

// Processes a single file
func processFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Fehler beim Öffnen der Datei:", err)
		return
	}
	defer file.Close()

	var content []string
	scanner := bufio.NewScanner(file)
	var insideRunde bool
	var block []string

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.Contains(line, "<runde>"):
			insideRunde = true
			block = []string{}
			continue

		case strings.Contains(line, "</runde>"):
			insideRunde = false
			markdownTable := convertRundeToMarkdown(strings.Join(block, "\n"))
			content = append(content, markdownTable)
			continue
		}

		if insideRunde {
			block = append(block, line)
		} else {
			content = append(content, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Fehler beim Lesen der Datei:", err)
		return
	}

	// Only save if there are changes
	originalContent, _ := os.ReadFile(path)
	newContent := strings.Join(content, "\n")

	if string(originalContent) != newContent {
		err = os.WriteFile(path, []byte(newContent), 0644)
		if err != nil {
			fmt.Println("Fehler beim Schreiben der Datei:", err)
		} else {
			fmt.Println("Paarungen in Datei umgewandelt:", path)
		}
	}
}

func main() {
	contentDir := "./content" // Path to the content folder
	processFiles(contentDir)
}
