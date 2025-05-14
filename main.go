package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Her domain için bulunan alt domainleri saklamak için yapı
type DomainResult struct {
	domain     string
	subdomains map[string]bool // Benzersiz alt domainleri tutmak için map
}

func main() {
	// Komut satırı bayraklarını tanımla
	var (
		inputFile string
		outputDir string
	)

	flag.StringVar(&inputFile, "f", "", "Domain listesi içeren dosya yolu")
	flag.StringVar(&inputFile, "file", "", "Domain listesi içeren dosya yolu (uzun format)")
	flag.StringVar(&outputDir, "o", "", "Sonuçların yazılacağı dizin (boş bırakılırsa mevcut dizin kullanılır)")
	flag.StringVar(&outputDir, "output", "", "Sonuçların yazılacağı dizin (uzun format)")

	// Kullanım bilgisi
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Kullanım: %s -f <dosya_yolu> [-o <çıktı_dizini>]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Parametreler:\n")
		flag.PrintDefaults()
	}

	// Bayrakları ayrıştır
	flag.Parse()

	// Zorunlu parametreleri kontrol et
	if inputFile == "" {
		fmt.Println("Hata: Dosya yolu belirtilmedi!")
		flag.Usage()
		os.Exit(1)
	}

	// Çıktı dizini belirtildiyse, dizinin varlığını kontrol et ve gerekirse oluştur
	if outputDir != "" {
		// Dizinin var olup olmadığını kontrol et
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			// Dizin yoksa oluştur
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				log.Fatalf("Çıktı dizini oluşturulamadı: %v", err)
			}
		}
	}

	// Dosyayı aç
	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("Dosya açılamadı: %v", err)
	}
	defer file.Close()

	// Tüm sonuçları tutacak slice
	var results []DomainResult

	// Dosyayı satır satır oku
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domain := strings.TrimSpace(scanner.Text())
		if domain == "" {
			continue
		}

		// Her domain için crt.sh'den veri çek
		fmt.Printf("Domain: %s için veri çekiliyor...\n", domain)
		result := scrapeData(domain)
		results = append(results, result)

		// Sonuçları dosyaya yaz
		writeResultToFile(result, outputDir)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Dosya okuma hatası: %v", err)
	}

	// İşlemler tamamlandıktan sonra tüm sonuçları yazdır
	fmt.Println("\n=== SONUÇLAR ===")
	for _, result := range results {
		fmt.Printf("\n%s domain adresi için bulunan benzersiz subdomainler:\n", result.domain)

		// Dosya yolunu oluştur
		fileName := result.domain + ".txt"
		filePath := fileName
		if outputDir != "" {
			filePath = filepath.Join(outputDir, fileName)
		}
		fmt.Printf("Sonuçlar %s dosyasına kaydedildi.\n", filePath)

		// Map'teki benzersiz alt domainleri yazdır
		if len(result.subdomains) == 0 {
			fmt.Println("Alt domain bulunamadı.")
		} else {
			for subdomain := range result.subdomains {
				fmt.Println("- " + subdomain)
			}
		}
	}
}

// Sonuçları belirtilen dizinde domain adıyla bir dosyaya yazar
func writeResultToFile(result DomainResult, outputDir string) {
	// Çıktı dosyası adını oluştur
	fileName := result.domain + ".txt"

	// Tam dosya yolunu oluştur
	filePath := fileName
	if outputDir != "" {
		filePath = filepath.Join(outputDir, fileName)
	}

	// Dosya yolunu güvenli hale getir
	filePath = filepath.Clean(filePath)

	// Dosyayı oluştur (varsa üzerine yaz)
	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("Dosya oluşturma hatası (%s): %v", filePath, err)
		return
	}
	defer file.Close()

	// Dosya başlığını yaz
	file.WriteString(fmt.Sprintf("# %s domain adresi için bulunan benzersiz subdomainler:\n\n", result.domain))

	// Subdomain sonuçlarını dosyaya yaz
	if len(result.subdomains) == 0 {
		file.WriteString("Alt domain bulunamadı.\n")
	} else {
		for subdomain := range result.subdomains {
			file.WriteString(subdomain + "\n")
		}
	}

	log.Printf("%s için sonuçlar %s dosyasına kaydedildi.", result.domain, filePath)
}

func scrapeData(domain string) DomainResult {
	// Sonuçları saklamak için yapı
	result := DomainResult{
		domain:     domain,
		subdomains: make(map[string]bool),
	}

	// crt.sh URL'sini oluştur
	url := fmt.Sprintf("https://crt.sh/?q=%s", domain)

	// HTTP isteği gönder
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("HTTP isteği hatası (%s): %v", domain, err)
		return result
	}
	defer resp.Body.Close()

	// HTML'i parse et
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("HTML parse hatası (%s): %v", domain, err)
		return result
	}

	// Tabloyu bul ve gerekli verileri çıkar
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		// tr içindeki tüm td'leri bul
		tds := s.Find("td")

		// Sondan bir önceki td elementini al
		if tds.Length() >= 2 {
			secondLastTd := tds.Eq(tds.Length() - 2)

			// HTML içeriğini al
			html, _ := secondLastTd.Html()

			// Farklı <br> formatlarını yakalamak için regex kullan
			brRegex := regexp.MustCompile(`<br\s*/?>\s*`)
			domainsText := brRegex.ReplaceAllString(html, "\n")

			// HTML etiketlerini temizle
			cleanDomainsText := removeHTMLTags(domainsText)

			// Alt domainleri satır satır ayır ve boşlukları temizle
			for _, subdomain := range strings.Split(cleanDomainsText, "\n") {
				subdomain = strings.TrimSpace(subdomain)
				if subdomain != "" {
					// Benzersiz alt domainleri ekle
					result.subdomains[subdomain] = true
				}
			}
		}
	})

	return result
}

func removeHTMLTags(html string) string {
	// HTML etiketlerini temizle
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)
	text := htmlTagRegex.ReplaceAllString(html, "")

	// HTML entity'leri düzelt (&amp; gibi)
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")

	return strings.TrimSpace(text)
}
