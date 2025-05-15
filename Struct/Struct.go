package Struct

type CertificateInfo struct {
	IssuerCAID     int    `json:"issuer_ca_id"`
	IssuerName     string `json:"issuer_name"`
	CommonName     string `json:"common_name"`
	NameValue      string `json:"name_value"`
	ID             int64  `json:"id"`
	EntryTimestamp string `json:"entry_timestamp"`
	NotBefore      string `json:"not_before"`
	NotAfter       string `json:"not_after"`
	SerialNumber   string `json:"serial_number"`
	ResultCount    int    `json:"result_count"`
}

type DomainResult struct {
	Domain     string
	Subdomains map[string]bool
}

type FlagStruct struct {
	InputFile  string
	Domain     string
	OutputFile string
}
