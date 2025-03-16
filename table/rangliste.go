package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Converts ranking list data into a Markdown table, ensuring correct column alignment
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

	// If there are no lines, return the original data
	if len(cleanLines) < 1 {
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
	for _, line := range cleanLines {
		columns := strings.Split(line, "\t")

		// Ensure there are enough columns before accessing
		if len(columns) < 12 {
			fmt.Println("Warning: Ungueltige Zeile uebersprungen (zu wenige Spalten):", line)
			continue
		}

		// Sicherstellen, dass die Spaltenstruktur erhalten bleibt
		verein := columns[5]
		land := columns[6]

		// Falls "Verein/Land" als ein Feld geschrieben wurde
		if strings.Contains(verein, "/") {
			parts := strings.SplitN(verein, "/", 2) // Maximal 2 Teile splitten
			verein = parts[0]
			land = parts[1]
		}

		// Falls "Land" noch leer ist, setzen wir ein leeres Feld, damit die Spaltenstruktur erhalten bleibt
		if land == "" {
			land = "-"
		}

		// Ersetze das urspruengliche Feld mit separatem Club & Land
		newColumns := append(columns[:5], verein, land)
		newColumns = append(newColumns, columns[7:]...) // Restliche Spalten wieder hinzufuegen

		// Append the formatted row
		mdLines = append(mdLines, "| "+strings.Join(newColumns, " | ")+" |")
	}

	return strings.Join(mdLines, "\n")
}

// Processes all .md files in the content directory
func processFiles(root string) {
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Println("Fehler beim Zugriff auf Datei:", err)
			return err
		}

		// Only process Markdown files
		if !d.IsDir() && filepath.Ext(path) == ".md" {
			processFile(path)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Fehler beim Verarbeiten der Dateien:", err)
	}
}

// Processes a single Markdown file and replaces ranking list blocks
func processFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Fehler beim Oeffnen der Datei:", err)
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
		case strings.Contains(line, "<rangliste>"):
			insideRangliste = true
			block = []string{}
			continue

		case strings.Contains(line, "</rangliste>"):
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
