package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Converts ranking list data into a Markdown table, separating Club and Country
func convertRanglisteToMarkdown(rangliste string) string {
	lines := strings.Split(rangliste, "\n")

	// Remove empty lines and trim spaces
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	// If there are fewer than 2 lines, return the original data
	if len(cleanLines) < 2 {
		return rangliste
	}

	// Define table headers (Club and Country are now separate)
	headers := []string{
		"Rang", "Teilnehmer", "Titel", "TWZ", "Attr.", "Verein", "Land", "S", "R", "V", "Punkte", "Buchh", "SoBerg",
	}

	// Create the Markdown table header
	mdLines := []string{
		"| " + strings.Join(headers, " | ") + " |",
		"| " + strings.Repeat("--- | ", len(headers)),
	}

	// Process each line and convert it into a table row
	for _, line := range cleanLines[1:] {
		columns := strings.Split(line, "\t")

		// Ensure there are enough columns by filling missing ones
		for len(columns) < len(headers)-1 {
			columns = append(columns, "")
		}

		// Split the Club and Country field
		clubAndCountry := strings.Split(columns[5], "/")
		club := ""
		country := ""
		if len(clubAndCountry) == 2 {
			club = clubAndCountry[0]
			country = clubAndCountry[1]
		} else if len(clubAndCountry) == 1 {
			club = clubAndCountry[0]
		}

		// Replace the combined field with separated Club and Country
		columns = append(columns[:5], append([]string{club, country}, columns[6:]...)...)

		// Append the formatted row
		mdLines = append(mdLines, "| "+strings.Join(columns, " | ")+" |")
	}

	return strings.Join(mdLines, "\n")
}

// Processes all .md files in the content directory
func processFiles(root string) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Fehler beim Zugriff auf Datei:", err)
		}

		// Only process Markdown files
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			processFile(path)
		}
		return nil
	})
}

// Processes a single Markdown file and replaces ranking list blocks
func processFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Fehler beim Öffnen der Datei:", err)
		return
	}
	defer file.Close()

	var content []string
	scanner := bufio.NewScanner(file)
	var insideRangliste bool
	var block []string

	// Read the file line by line
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.Contains(line, "{{< rangliste >}}"):
			insideRangliste = true
			block = []string{}
			continue

		case strings.Contains(line, "{{< /rangliste >}}"):
			insideRangliste = false
			// Convert ranking list block to Markdown table
			markdownTable := convertRanglisteToMarkdown(strings.Join(block, "\n"))
			content = append(content, markdownTable)
			continue
		}

		if insideRangliste {
			block = append(block, line)
		} else {
			content = append(content, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Fehler beim Lesen der Datei:", err)
		return
	}

	// Only overwrite the file if content has changed
	originalContent, _ := os.ReadFile(path)
	newContent := strings.Join(content, "\n")

	if string(originalContent) != newContent {
		err = os.WriteFile(path, []byte(newContent), 0644)
		if err != nil {
			fmt.Println("Fehler beim Schreiben der Datei:", err)
		} else {
			fmt.Println("Rangliste erfolgreich konvertiert:", path)
		}
	}
}

func main() {
	contentDir := "./content" // Path to the content directory
	processFiles(contentDir)
}
