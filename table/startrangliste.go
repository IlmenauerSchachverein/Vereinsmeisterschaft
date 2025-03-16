package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// convertRanglisteToMarkdown wandelt einen <rangliste> Block in eine Markdown-Tabelle um.
func convertRanglisteToMarkdown(rangliste string) string {
	lines := strings.Split(rangliste, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}
	if len(cleanLines) < 1 {
		return rangliste
	}
	headers := []string{"Rang", "Teilnehmer", "Titel", "TWZ", "Attr.", "Verein", "Land", "S", "R", "V", "Punkte", "Buchh", "SoBerg"}
	mdLines := []string{
		"| " + strings.Join(headers, " | ") + " |",
		"| " + strings.Repeat("--- | ", len(headers)),
	}
	for _, line := range cleanLines {
		columns := strings.Split(line, "\t")
		if len(columns) < 12 {
			// Fehlende Spalten mit leeren Strings auffüllen
			diff := 12 - len(columns)
			for i := 0; i < diff; i++ {
				columns = append(columns, "")
			}
		}
		verein := columns[5]
		land := columns[6]
		if strings.Contains(verein, "/") {
			parts := strings.SplitN(verein, "/", 2)
			verein = parts[0]
			land = parts[1]
		}
		if land == "" {
			land = "-"
		}
		newColumns := append(columns[:5], verein, land)
		newColumns = append(newColumns, columns[7:]...)
		mdLines = append(mdLines, "| "+strings.Join(newColumns, " | ")+" |")
	}
	return strings.Join(mdLines, "\n")
}

// convertStartranglisteToMarkdown wandelt einen <startrangliste> Block in eine Markdown-Tabelle um.
// Da der Header nicht im Input vorhanden ist, wird er hier fest vorgegeben.
func convertStartranglisteToMarkdown(startrangliste string) string {
	lines := strings.Split(startrangliste, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}
	headers := []string{"TlnNr", "Teilnehmer", "Titel", "ELO", "NWZ", "Attr.", "Verein/Ort", "Land", "Geburt", "FideKenn.", "PKZ"}
	mdLines := []string{
		"| " + strings.Join(headers, " | ") + " |",
		"| " + strings.Repeat("--- | ", len(headers)),
	}
	// Alle Zeilen als Datensätze behandeln
	for _, line := range cleanLines {
		columns := strings.Split(line, "\t")
		if len(columns) < len(headers) {
			// Fehlende Spalten mit leeren Strings auffüllen
			diff := len(headers) - len(columns)
			for i := 0; i < diff; i++ {
				columns = append(columns, "")
			}
		}
		mdLines = append(mdLines, "| "+strings.Join(columns, " | ")+" |")
	}
	return strings.Join(mdLines, "\n")
}

// processFiles durchläuft alle .md Dateien im angegebenen Verzeichnis.
func processFiles(root string) {
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Println("Fehler beim Zugriff auf Datei:", err)
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".md" {
			processFile(path)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Fehler beim Verarbeiten der Dateien:", err)
	}
}

// processFile bearbeitet eine einzelne Markdown-Datei und ersetzt <rangliste> und <startrangliste> Blöcke.
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
	var blockRangliste []string

	var insideStartrangliste bool
	var blockStartrangliste []string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "<startrangliste>") {
			insideStartrangliste = true
			blockStartrangliste = []string{}
			continue
		}
		if strings.Contains(line, "</startrangliste>") {
			insideStartrangliste = false
			markdownTable := convertStartranglisteToMarkdown(strings.Join(blockStartrangliste, "\n"))
			content = append(content, markdownTable)
			continue
		}
		if insideStartrangliste {
			blockStartrangliste = append(blockStartrangliste, line)
			continue
		}
		if strings.Contains(line, "<rangliste>") {
			insideRangliste = true
			blockRangliste = []string{}
			continue
		}
		if strings.Contains(line, "</rangliste>") {
			insideRangliste = false
			markdownTable := convertRanglisteToMarkdown(strings.Join(blockRangliste, "\n"))
			content = append(content, markdownTable)
			continue
		}
		if insideRangliste {
			blockRangliste = append(blockRangliste, line)
			continue
		}
		content = append(content, line)
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Fehler beim Lesen der Datei:", err)
		return
	}
	originalContent, _ := os.ReadFile(path)
	newContent := strings.Join(content, "\n")
	if string(originalContent) != newContent {
		err = os.WriteFile(path, []byte(newContent), 0644)
		if err != nil {
			fmt.Println("Fehler beim Schreiben der Datei:", err)
		} else {
			fmt.Println("Datei erfolgreich aktualisiert:", path)
		}
	}
}

func main() {
	contentDir := "./content"
	processFiles(contentDir)
}
