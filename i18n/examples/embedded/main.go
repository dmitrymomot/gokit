package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"

	"github.com/dmitrymomot/gokit/i18n"
)

//go:embed translations/*.yaml
var translationsFS embed.FS

func main() {
	// Load translations from embedded filesystem
	httpFS := http.FS(translationsFS)
	err := i18n.LoadTranslationsFS(httpFS, "translations")
	if err != nil {
		log.Fatalf("Failed to load translations: %v", err)
	}

	// Test different languages
	languages := []string{"en", "fr"}
	
	for _, lang := range languages {
		fmt.Printf("\n=== %s ===\n", lang)
		fmt.Println(i18n.T(lang, "welcome", "name", "User"))
		fmt.Println(i18n.T(lang, "about"))
		
		// Test pluralization
		for _, count := range []int{0, 1, 5} {
			countStr := fmt.Sprintf("%d", count)
			fmt.Println(i18n.N(lang, "status", count, "count", countStr))
		}
	}
	
	// Test language negotiation
	acceptLang := "fr-CA,fr;q=0.9,en-US;q=0.8,en;q=0.7"
	fmt.Printf("\n=== Accept-Language: %s ===\n", acceptLang)
	bestLang := i18n.BestLangFromAcceptLanguage(acceptLang)
	fmt.Printf("Best language: %s\n", bestLang)
	fmt.Println(i18n.T(bestLang, "welcome", "name", "User"))
}
