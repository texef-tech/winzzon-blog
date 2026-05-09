package main

import (
	"fmt"
	"net/http"

	"github.com/texef-tech/winzzon-blog/internal/config"
)

func main() {
	cfg := config.Load()

	fmt.Printf("Starting server on port %s (env: %s)\n", cfg.Port, cfg.Env)

	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		panic(err)
	}
}
