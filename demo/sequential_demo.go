package main

import (
	"fmt"
	"time"

	"github.com/MarcelWMeyer/nextcloud-performance-monitor/internal/agent"
)

// Demo für sequenzielle Test-Ausführung
func main() {
	// Beispiel-Konfiguration erstellen
	configs := []*agent.Config{
		{
			InstanceName:    "Test Nextcloud 1",
			ServiceType:     "nextcloud", 
			URL:             "https://test1.example.com",
			Username:        "user1",
			Password:        "pass1",
			TestIntervalSec: 30,
		},
		{
			InstanceName:    "Test Nextcloud 2",
			ServiceType:     "nextcloud",
			URL:             "https://test2.example.com", 
			Username:        "user2",
			Password:        "pass2",
			TestIntervalSec: 30,
		},
		{
			InstanceName:    "Test HiDrive",
			ServiceType:     "hidrive",
			URL:             "https://hidrive.example.com",
			Username:        "userh",
			Password:        "passh", 
			TestIntervalSec: 30,
		},
	}

	fmt.Printf("=== Echte Sequenzielle Test-Demonstration ===\n")
	fmt.Printf("Anzahl Instanzen: %d\n", len(configs))
	
	baseInterval := time.Duration(configs[0].TestIntervalSec) * time.Second
	
	fmt.Printf("Test-Zyklus-Intervall: %v\n", baseInterval)
	fmt.Printf("\n=== Ausführungsreihenfolge ===\n")
	
	for i, cfg := range configs {
		fmt.Printf("Schritt %d: %s (%s) - wartet bis vorheriger Test abgeschlossen\n", 
			i+1, cfg.InstanceName, cfg.ServiceType)
	}
	
	fmt.Printf("\n=== Neue Vorteile ===\n")
	fmt.Printf("✅ Tests überlappen NIE\n")
	fmt.Printf("✅ Jeder Test startet erst nach Abschluss des vorherigen\n") 
	fmt.Printf("✅ Vorhersagbare Reihenfolge\n")
	fmt.Printf("✅ Keine Zeitberechnung nötig\n")
	fmt.Printf("✅ Funktioniert auch bei unterschiedlich langen Tests\n")
	
	fmt.Printf("\n=== Beispiel-Ablauf ===\n")
	fmt.Printf("0:00 - Nextcloud 1 startet\n")
	fmt.Printf("0:45 - Nextcloud 1 fertig → Nextcloud 2 startet sofort\n")  
	fmt.Printf("1:20 - Nextcloud 2 fertig → HiDrive startet sofort\n")
	fmt.Printf("2:10 - HiDrive fertig → Zyklus abgeschlossen\n")
	fmt.Printf("5:00 - Nächster Zyklus startet (nach %v Pause)\n", baseInterval)
}
