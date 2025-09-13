package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// AlertWebhook represents the structure of an Alertmanager webhook
type AlertWebhook struct {
	Version           string                 `json:"version"`
	GroupKey          string                 `json:"groupKey"`
	Status            string                 `json:"status"`
	Receiver          string                 `json:"receiver"`
	GroupLabels       map[string]string      `json:"groupLabels"`
	CommonLabels      map[string]string      `json:"commonLabels"`
	CommonAnnotations map[string]string      `json:"commonAnnotations"`
	ExternalURL       string                 `json:"externalURL"`
	Alerts            []Alert                `json:"alerts"`
}

type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var webhook AlertWebhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	// Log the alert
	log.Printf("🔔 ALERT RECEIVED [%s] - Receiver: %s", webhook.Status, webhook.Receiver)
	log.Printf("   Group: %s", webhook.GroupKey)
	
	for i, alert := range webhook.Alerts {
		severity := alert.Labels["severity"]
		category := alert.Labels["category"]
		instance := alert.Labels["instance"]
		
		var emoji string
		switch severity {
		case "critical":
			emoji = "🚨"
		case "warning":
			emoji = "⚠️"
		default:
			emoji = "ℹ️"
		}
		
		log.Printf("   %s Alert %d/%d:", emoji, i+1, len(webhook.Alerts))
		log.Printf("     Name: %s", alert.Labels["alertname"])
		log.Printf("     Severity: %s", severity)
		log.Printf("     Category: %s", category)
		log.Printf("     Instance: %s", instance)
		log.Printf("     Summary: %s", alert.Annotations["summary"])
		log.Printf("     Description: %s", alert.Annotations["description"])
		if runbook := alert.Annotations["runbook_url"]; runbook != "" {
			log.Printf("     Runbook: %s", runbook)
		}
		log.Printf("     Status: %s", alert.Status)
		if !alert.StartsAt.IsZero() {
			log.Printf("     Started: %s", alert.StartsAt.Format("2006-01-02 15:04:05"))
		}
		if !alert.EndsAt.IsZero() && alert.Status == "resolved" {
			duration := alert.EndsAt.Sub(alert.StartsAt)
			log.Printf("     Resolved: %s (Duration: %s)", alert.EndsAt.Format("2006-01-02 15:04:05"), duration)
		}
	}
	
	log.Println("   ─────────────────────────────────────")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Alert received successfully")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Webhook logger is healthy")
}

func main() {
	http.HandleFunc("/webhook", webhookHandler)
	http.HandleFunc("/webhook/critical", webhookHandler)
	http.HandleFunc("/health", healthHandler)
	
	log.Println("🔔 Webhook Logger starting on :8080")
	log.Println("📍 Endpoints:")
	log.Println("   • POST /webhook - General alerts")
	log.Println("   • POST /webhook/critical - Critical alerts")
	log.Println("   • GET  /health - Health check")
	log.Println("─────────────────────────────────────")
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
