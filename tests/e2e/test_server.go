package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/coinbase/x402/go/mechanisms/multiversx/exact/facilitator"
	"github.com/coinbase/x402/go/types"
)

func main() {
	apiUrl := os.Getenv("MULTIVERSX_API_URL")
	if apiUrl == "" {
		apiUrl = "https://devnet-gateway.multiversx.com"
	}

	verifier := facilitator.NewExactMultiversXScheme(apiUrl)
	// serverScheme := server.NewExactMultiversXScheme() // unused

	// Create a simple handler that performs verification
	http.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Payload      types.PaymentPayload      `json:"payload"`
			Requirements types.PaymentRequirements `json:"requirements"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Payload is already correct type
		payload := req.Payload

		// Verify
		resp, err := verifier.Verify(r.Context(), payload, req.Requirements)
		if err != nil {
			log.Printf("Verification failed: %v", err)
			http.Error(w, fmt.Sprintf("Verification failed: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Test Server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
