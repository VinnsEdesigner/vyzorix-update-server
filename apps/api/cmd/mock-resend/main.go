// Mock Resend API server for testing email functionality.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

var (
	sentEmails []EmailLog
	port       = "1080"
)

type EmailLog struct { //nolint:govet // fieldalignment issue is acceptable for mock server
	ID          string    `json:"id"`
	To          []string  `json:"to"`
	From        string    `json:"from"`
	Subject     string    `json:"subject"`
	HTML        string    `json:"html"`
	Timestamp   time.Time `json:"timestamp"`
	VerifyToken string    `json:"verify_token,omitempty"`
	VerifyURL   string    `json:"verify_url,omitempty"`
}

func main() {
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	dir, _ := os.Getwd()
	logFile, err := os.OpenFile(dir+"/emails.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)

	log.Printf("Mock Resend server starting on :%s\n", port)

	http.HandleFunc("/emails", handleSendEmail)
	http.HandleFunc("/emails/", handleGetEmail)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/latest", handleGetLatestEmail)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("encode error: %v", err)
	}
}

func handleSendEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	html := toString(payload["html"])

	// Extract verification token from HTML.
	token, verifyURL := extractVerificationToken(html)

	email := EmailLog{
		ID:          fmt.Sprintf("mock_%d", time.Now().UnixNano()),
		To:          toStringArray(payload["to"]),
		From:        toString(payload["from"]),
		Subject:     toString(payload["subject"]),
		HTML:        html,
		Timestamp:   time.Now(),
		VerifyToken: token,
		VerifyURL:   verifyURL,
	}

	sentEmails = append(sentEmails, email)

	// Save to JSON file for easy parsing.
	saveEmailsToFile()

	// Log to file.
	log.Printf("=== EMAIL SENT ===")
	log.Printf("ID: %s", email.ID)
	log.Printf("To: %v", email.To)
	log.Printf("From: %s", email.From)
	log.Printf("Subject: %s", email.Subject)
	log.Printf("Time: %s", email.Timestamp.Format(time.RFC3339))
	if token != "" {
		log.Printf("Verify Token: %s", token)
		log.Printf("Verify URL: %s", verifyURL)
	}
	if html := email.HTML; len(html) > 200 {
		log.Printf("HTML (preview): %s...", html[:200])
	}
	log.Printf("=================")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"id":   email.ID,
		"from": email.From,
		"to":   email.To,
	}); err != nil {
		log.Printf("encode error: %v", err)
	}
}

func handleGetEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sentEmails); err != nil {
		log.Printf("encode error: %v", err)
	}
}

func handleGetLatestEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if len(sentEmails) > 0 {
		if err := json.NewEncoder(w).Encode(sentEmails[len(sentEmails)-1]); err != nil {
			log.Printf("encode error: %v", err)
		}
	} else {
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "no emails sent"}); err != nil {
			log.Printf("encode error: %v", err)
		}
	}
}

func saveEmailsToFile() {
	dir, _ := os.Getwd()
	data, err := json.MarshalIndent(sentEmails, "", "  ")
	if err != nil {
		log.Printf("marshal error: %v", err)
		return
	}
	if err := os.WriteFile(dir+"/emails.json", data, 0644); err != nil {
		log.Printf("write file error: %v", err)
	}
}

func extractVerificationToken(html string) (string, string) {
	// Match URLs with token parameter.
	re := regexp.MustCompile(`([?&]token=)([^&"'\s]+)`)
	matches := re.FindStringSubmatch(html)
	if len(matches) >= 3 {
		token := matches[2]
		// Find the full URL.
		urlRe := regexp.MustCompile(`https?://[^"'\s]+[?&]token=` + regexp.QuoteMeta(token) + `[^"'\s]*`)
		urlMatch := urlRe.FindString(html)
		return token, urlMatch
	}
	return "", ""
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func toStringArray(v interface{}) []string {
	if arr, ok := v.([]interface{}); ok {
		result := make([]string, len(arr))
		for i, item := range arr {
			if s, ok := item.(string); ok {
				result[i] = s
			}
		}
		return result
	}
	return nil
}
