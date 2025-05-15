package Functions

import (
	"ListCRTSH/Struct"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func FlagParse() Struct.FlagStruct {
	var inputFile, outputDir, domain string

	flag.StringVar(&inputFile, "f", "", "Domain listesi içeren dosya yolu")
	flag.StringVar(&domain, "d", "", "Domain adı")
	flag.StringVar(&outputDir, "o", "", "Sonuçların yazılacağı dizin (boş bırakılırsa mevcut dizin kullanılır)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Kullanım: %s -f <dosya_yolu> [-o <çıktı_dizini>]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Parametreler:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if inputFile == "" && domain == "" {
		fmt.Println("Hata: Dosya yolu ya da domain adresi belirtilmedi!")
		flag.Usage()
		os.Exit(1)
	}

	if outputDir != "" {
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				log.Fatalf("Çıktı dizini oluşturulamadı: %v", err)
			}
		}
	}

	return Struct.FlagStruct{
		InputFile:  inputFile,
		Domain:     domain,
		OutputFile: outputDir,
	}
}

func ScrapeData(domain string) Struct.DomainResult {

	result := Struct.DomainResult{
		Domain:     domain,
		Subdomains: make(map[string]bool),
	}
	url := fmt.Sprintf("https://crt.sh/json?q=%s", domain)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("HTTP isteği hatası (%s): %v", domain, err)
		return result
	}

	respByte, err := io.ReadAll(resp.Body)
	var CertificateInfo []Struct.CertificateInfo
	json.Unmarshal(respByte, &CertificateInfo)

	for _, info := range CertificateInfo {
		result.Subdomains[info.CommonName] = true
	}
	defer resp.Body.Close()

	return result
}

func WriteResultToFile(result Struct.DomainResult, outputDir string) {

	fileName := result.Domain + ".txt"

	filePath := fileName
	if outputDir != "" {
		filePath = filepath.Join(outputDir, fileName)
	}

	filePath = filepath.Clean(filePath)

	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("Dosya oluşturma hatası (%s): %v", filePath, err)
		return
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("# %s domain adresi için bulunan benzersiz subdomainler:\n\n", result.Domain))

	if len(result.Subdomains) == 0 {
		file.WriteString("Alt domain bulunamadı.\n")
	} else {
		for subdomain := range result.Subdomains {
			file.WriteString(subdomain + "\n")
		}
	}

	log.Printf("%s için sonuçlar %s dosyasına kaydedildi.", result.Domain, filePath)
}

func TerminalOutput(results Struct.DomainResult) {
	if len(results.Subdomains) == 0 {
		fmt.Println("Alt domain bulunamadı.")
	} else {
		for result, _ := range results.Subdomains {
			fmt.Println(result)
		}
		fmt.Println("*********************")
	}
}
