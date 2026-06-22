package rag

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ExtractText extracts text content from various file formats
func ExtractText(filename string, data []byte, fileType string) (string, error) {
	switch fileType {
	case "text/plain", "text/markdown", ".txt", ".md":
		return string(data), nil
	case "application/pdf", ".pdf":
		return extractPDF(data)
	default:
		// Try to interpret as plain text
		text := string(data)
		if isReadable(text) {
			return text, nil
		}
		return "", fmt.Errorf("unsupported file type: %s", fileType)
	}
}

func extractPDF(data []byte) (string, error) {
	reader := bytes.NewReader(data)
	pdfReader, err := pdf.NewReader(reader, int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to parse PDF: %w", err)
	}

	var sb strings.Builder
	for i := 1; i <= pdfReader.NumPage(); i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			// Skip pages we can't read
			continue
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}

	result := sb.String()
	if strings.TrimSpace(result) == "" {
		return "", fmt.Errorf("PDF appears to be empty or image-only (no extractable text)")
	}
	return result, nil
}

// isReadable checks if byte slice looks like readable text
func isReadable(s string) bool {
	printable := 0
	total := len(s)
	if total == 0 {
		return false
	}
	if total > 1000 {
		total = 1000
	}
	for i := 0; i < total; i++ {
		c := s[i]
		if c >= 32 || c == '\n' || c == '\r' || c == '\t' {
			printable++
		}
	}
	return float64(printable)/float64(total) > 0.85
}

// DetectFileType determines the file type from filename and content
func DetectFileType(filename string, _ io.Reader) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".pdf"):
		return ".pdf"
	case strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".markdown"):
		return ".md"
	case strings.HasSuffix(lower, ".txt"):
		return ".txt"
	default:
		return ".txt"
	}
}
