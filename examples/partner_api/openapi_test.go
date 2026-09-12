package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestEveryOpenAPIOperationHasSuccessResponseExample(t *testing.T) {
	raw, err := os.ReadFile("openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Paths map[string]map[string]struct {
			Responses map[string]struct {
				Content map[string]struct {
					Example any `json:"example"`
				} `json:"content"`
			} `json:"responses"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("openapi.json tidak valid: %v", err)
	}

	operationCount := 0
	for path, pathItem := range document.Paths {
		for method, operation := range pathItem {
			operationCount++
			success, ok := operation.Responses["200"]
			if !ok {
				t.Errorf("%s %s tidak mendokumentasikan response 200", method, path)
				continue
			}
			mediaType, ok := success.Content["application/json"]
			if !ok || mediaType.Example == nil {
				t.Errorf("%s %s tidak memiliki contoh response JSON", method, path)
			}
		}
	}
	if operationCount != 13 {
		t.Fatalf("jumlah operasi terdokumentasi = %d, ingin 13", operationCount)
	}
}
