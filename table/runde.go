package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// insertLineBreakAfterFirstCommaIfLong fügt einen Zeilenumbruch nach dem ersten Komma ein,
// falls der übergebene String länger als den angegebenen Schwellenwert ist.
func insertLineBreakAfterFirstCommaIfLong(s string, threshold int) string {
	if len(s) < threshold {
		return s
	}
	idx := strings.Index(s, ",")
	if idx == -1 {
		return s
	}
	// Optional: Sollte der Markdown-Renderer kein "\n" in Tabellenzellen verarbeiten,
	// kann hier auch "<br>" verwendet werden.
	return s[:idx+1] + "\n" + strings.TrimSpace(s[idx+1:])
}

// convertRundeToMarkdown wandelt einen <runde> Block in eine Markdown-Tabelle um.
// Es werden 14 Spalten erwartet, wobei:
// - Linke Seite: Spalten 0 bis 5 (Tisch, TNr, Teilnehmer, Titel, Punkte, -)
// - Rechte Seite: Spalten 6 bis 9 (TNr, Teilnehmer, Titel, Punkte)
// - Die Spalten 10, 11 und 12 werden zu einer Spalte "Ergebnis" zusammengeführt.
// - Spalte 13 wird verworfen.
// Die "Titel"-Spalten (Spalte 3 und Spalte 8) werden aus der Ausgabe entfernt.
// Zusätzlich wird bei langen Namen (länger als der definierte Schwellenwert)
// ein Zeilenumbruch nach dem ersten Komma eingefügt.
func convertRundeToMarkdown(runde string) string {
	lines := strings.Split(runde, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	// Neuer Header ohne die "Titel"-Spalten
	headers := []string{"Tisch", "TNr", "Teilnehmer", "Punkte", "-", "TNr", "Teilnehmer", "Punkte", "Ergebnis"}
	mdLines := []string{
		"| " + strings.Join(headers, " | ") + " |",
		"| " + strings.Repeat("----- | ", len(headers)),
	}

	// Definiere den Schwellenwert für lange Namen
	const nameThreshold = 18

	for _, line := range cleanLines {
		// Aufspalten der Zeile anhand von Tabulatoren
		columns := strings.Split(line, "\t")
		// Falls weniger als 14 Spalten vorhanden, mit leeren Strings auffüllen
		for len(columns) < 14 {
			columns = append(columns, "")
		}
		if len(columns) < 14 {
			fmt.Println("Warning: Ungültige Zeile übersprungen (zu wenige Spalten):", line)
			continue
		}

		// Zusammenführen der Spalten 10, 11 und 12 zu einer Spalte "Ergebnis"
		merged := strings.TrimSpace(columns[10]) + " " + strings.TrimSpace(columns[11]) + " " + strings.TrimSpace(columns[12])
		merged = strings.Join(strings.Fields(merged), " ") // entfernt überflüssige Leerzeichen

		// Anwenden des Zeilenumbruchs bei langen Namen in den Teilnehmer-Spalten
		teilnehmerLinks := insertLineBreakAfterFirstCommaIfLong(strings.TrimSpace(columns[2]), nameThreshold)
		teilnehmerRechts := insertLineBreakAfterFirstCommaIfLong(strings.TrimSpace(columns[7]), nameThreshold)

		// Neue Spalten: Entferne "Titel"-Spalten (Spalte 3 und Spalte 8)
		newColumns := []string{
			strings.TrimSpace(columns[0]), // Tisch
			strings.TrimSpace(columns[1]), // TNr
			teilnehmerLinks,               // Teilnehmer (links) mit ggf. Zeilenumbruch
			strings.TrimSpace(columns[4]), // Punkte (anstelle von Titel)
			strings.TrimSpace(columns[5]), // -
			strings.TrimSpace(columns[6]), // TNr (rechte Seite)
			teilnehmerRechts,              // Teilnehmer (rechts) mit ggf. Zeilenumbruch
			strings.TrimSpace(columns[9]), // Punkte (rechte Seite, statt Titel)
			merged,                        // Ergebnis (gemergt)
		}

		mdLines = append(mdLines, "| "+strings.Join(newColumns, " | ")+" |")
	}

	return strings.Join(mdLines, "\n")
}

// processFiles durchsucht rekursiv alle .md Dateien im angegebenen Verzeichnis.
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

// processFile bearbeitet eine einzelne Markdown-Datei und ersetzt <runde> Blöcke.
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
	var blockRunde []string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "<runde>") {
			insideRunde = true
			blockRunde = []string{}
			continue
		}
		if strings.Contains(line, "</runde>") {
			insideRunde = false
			mdTable := convertRundeToMarkdown(strings.Join(blockRunde, "\n"))
			content = append(content, mdTable)
			continue
		}
		if insideRunde {
			blockRunde = append(blockRunde, line)
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
