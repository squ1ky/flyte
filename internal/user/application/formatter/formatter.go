package formatter

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
	"unicode"
)

// FormatPhoneNumber applies upper case and leaves only letters and digits
// Example: "1234 567890" -> "1234567890"
func FormatPhoneNumber(phone string) string {
	var sb strings.Builder
	for _, r := range phone {
		if unicode.IsDigit(r) || r == '+' {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// FormatDocumentNumber applies upper case and removes spaces
// Example: " 12 34 567890 " -> "1234567890"
func FormatDocumentNumber(docNum string) string {
	return strings.ToUpper(strings.Join(strings.Fields(docNum), ""))
}

// FormatDocumentType applies lower case for document type
// Example: "Passport" -> "passport"
func FormatDocumentType(docType string) string {
	return strings.ToLower(strings.TrimSpace(docType))
}

// FormatName applies title case and removes spaces
// Example: "   iVaN  " -> "Ivan"
func FormatName(name string) string {
	name = strings.TrimSpace(name)
	return cases.Title(language.Und).String(strings.ToLower(name))
}
