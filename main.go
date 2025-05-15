package main

import (
	"ListCRTSH/Functions"
	"ListCRTSH/Struct"
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {

	var FlagStruct Struct.FlagStruct = Functions.FlagParse()
	var results Struct.DomainResult

	if FlagStruct.Domain != "" {
		results = Functions.ScrapeData(strings.TrimSpace(FlagStruct.Domain))
	} else if FlagStruct.InputFile != "" {

		file, err := os.Open(FlagStruct.InputFile)
		if err != nil {
			log.Fatalf("Dosya açılamadı: %v", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		fmt.Println("****** SONUÇLAR ******")
		for scanner.Scan() {
			domain := strings.TrimSpace(scanner.Text())
			if domain == "" {
				continue
			}

			fmt.Printf("Domain: %s için veri çekiliyor...\n", domain)
			results = Functions.ScrapeData(domain)

			if FlagStruct.OutputFile != "" {
				Functions.WriteResultToFile(results, FlagStruct.OutputFile)
			}

			Functions.TerminalOutput(results)
		}

		if err := scanner.Err(); err != nil {
			log.Fatalf("Dosya okuma hatası: %v", err)
		}
	}
}
