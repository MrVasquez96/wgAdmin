package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Configuration: Change this if your English file has a different name
	sourceFile := "translations_en.go"

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter new language code (e.g., 'no-nb', 'fr'): ")
	langCode, _ := reader.ReadString('\n')
	langCode = strings.TrimRight(langCode, "\r\n")

	fmt.Print("Enter variable prefix (e.g., 'norwegian', 'french'): ")
	langPrefix, _ := reader.ReadString('\n')
	langPrefix = strings.TrimRight(langPrefix, "\r\n")

	// 1. Open the English source file
	file, err := os.Open(sourceFile)
	if err != nil {
		fmt.Printf("Error: Could not open %s. Make sure you are running this in the correct directory.\n", sourceFile)
		return
	}
	defer file.Close()

	// 2. Create the new target file
	outFile := fmt.Sprintf("translations_%s.go", langCode)
	out, err := os.Create(outFile)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", outFile, err)
		return
	}
	defer out.Close()

	// 1. Open the English source file
	tmp_file, err := os.Open(sourceFile)
	if err != nil {
		fmt.Printf("Error: Could not open %s. Make sure you are running this in the correct directory.\n", sourceFile)
		return
	}
	scanner_tmp := bufio.NewScanner(tmp_file)
	inMapw:= false
	count:=0
	for scanner_tmp.Scan() {
		line := scanner_tmp.Text()
		// Look for the start of the English map
		if strings.Contains(line, "var englishTranslations = Translations{") {
			inMapw = true
			// Write the file header for the new language file
			out.WriteString("package lang\n\n")
			out.WriteString(fmt.Sprintf("var %sTranslations = Translations{\n", langPrefix))
			continue
		}

		if !inMapw {
			// Skip everything above the map (type definitions, global vars, etc.)
			continue
		}

		// Look for the end of the map
		if strings.HasPrefix(strings.TrimSpace(line), "}") {
			out.WriteString("}\n")
			break 
		}
		if strings.Contains(line, ":") {
				count++
		}

	}
	tmp_file.Close()
	
	scanner := bufio.NewScanner(file)
	inMap := false
	translation_num := 0
	for scanner.Scan() {
		line := scanner.Text()

		// Look for the start of the English map
		if strings.Contains(line, "var englishTranslations = Translations{") {
			inMap = true
			// Write the file header for the new language file
			out.WriteString("package lang\n\n")
			out.WriteString(fmt.Sprintf("var %sTranslations = Translations{\n", langPrefix))
			continue
		}

		if !inMap {
			// Skip everything above the map (type definitions, global vars, etc.)
			continue
		}

		// Look for the end of the map
		if strings.HasPrefix(strings.TrimSpace(line), "}") {
			out.WriteString("}\n")
			break 
		}

		trimLine := strings.TrimSpace(line)

		// Preserve empty lines and comments automatically
		if trimLine == "" || strings.HasPrefix(trimLine, "//") {
			out.WriteString(line + "\n")
			continue
		}

		// Parse the key and the English string
		firstQuote := strings.Index(line, "\"")
		lastQuote := strings.LastIndex(line, "\"")

		if firstQuote != -1 && lastQuote != -1 && firstQuote != lastQuote {
			englishText := line[firstQuote+1 : lastQuote]
			prefix := line[:firstQuote]   // The key and indentation (e.g., "    Common_save:   ")
			suffix := line[lastQuote+1:]  // The trailing comma and any inline comments
			translation_num++
			fmt.Printf("\n[%d/%d]English: %s\n",translation_num,count ,englishText)
			fmt.Print("Translation (leave empty to keep English): ")
			
			translation, _ := reader.ReadString('\n')
			translation = strings.TrimRight(translation, "\r\n")

			if translation == "" {
				translation = englishText
			}

			// Write the newly translated line maintaining exact original spacing
			out.WriteString(fmt.Sprintf("%s\"%s\"%s\n", prefix, translation, suffix))
		} else {
			// Fallback: if the line doesn't match string patterns, copy it directly
			out.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}

	fmt.Printf("\nTranslations file generated: '%s'.\n", outFile)
	fmt.Printf("Register %q in types.go to use it.\n", langPrefix)
}