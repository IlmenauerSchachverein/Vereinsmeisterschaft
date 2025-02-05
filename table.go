package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Funktion zum Konvertieren eines Ranglisten-Blocks in eine Markdown-Tabelle
func convertToMarkdown(rangliste string) string {
	lines := strings.Split(rangliste, "\n")

	// Entferne leere Zeilen und trimme Leerzeichen
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	if len(cleanLines) < 2 {
		return rangliste // Nicht genug Daten, um eine Tabelle zu erstellen
	}

	// Header und Datenzeilen trennen (Tab als Trennzeichen)
	headers := strings.Split(cleanLines[1], "\t")
	separator := "| " + strings.Repeat("--- | ", len(headers))

	// Markdown-Header und Separator erstellen
	mdLines := []string{
		"| " + strings.Join(headers, " | ") + " |",
		separator,
	}

	// Datenzeilen verarbeiten
	for _, line := range cleanLines[2:] {
		columns := strings.Split(line, "\t")
		// Fehlende Spalten auffüllen
		for len(columns) < len(headers) {
			columns = append(columns, "")
		}
		mdLines = append(mdLines, "| "+strings.Join(columns, " | ")+" |")
	}

	return strings.Join(mdLines, "\n")
}

// Funktion zum Verarbeiten aller Dateien im content-Ordner
func processFiles(root string) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Fehler beim Zugriff auf Datei:", err)
			return nil
		}

		// Nur Markdown-Dateien verarbeiten
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			processFile(path)
		}
		return nil
	})
}

// Funktion zum Verarbeiten einer einzelnen Datei
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
	var ranglisteBlock []string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "{{< rangliste >}}") {
			insideRangliste = true
			ranglisteBlock = []string{}
			continue
		}

		if strings.Contains(line, "{{< /rangliste >}}") {
			insideRangliste = false
			// Konvertiere den Ranglisten-Block und füge ihn ein
			markdownTable := convertToMarkdown(strings.Join(ranglisteBlock, "\n"))
			content = append(content, markdownTable)
			continue
		}

		if insideRangliste {
			ranglisteBlock = append(ranglisteBlock, line)
		} else {
			content = append(content, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Fehler beim Lesen der Datei:", err)
		return
	}

	// Datei mit der neuen Tabelle überschreiben
	err = os.WriteFile(path, []byte(strings.Join(content, "\n")), 0644)
	if err != nil {
		fmt.Println("Fehler beim Schreiben der Datei:", err)
	} else {
		fmt.Println("Rangliste in Datei umgewandelt:", path)
	}
}

func main() {
	contentDir := "./content" // Pfad zum Content-Ordner
	processFiles(contentDir)
}
