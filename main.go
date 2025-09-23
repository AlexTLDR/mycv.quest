package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AlexTLDR/mycv.quest/pkg/config"
	"github.com/AlexTLDR/mycv.quest/pkg/generator"
	"github.com/AlexTLDR/mycv.quest/pkg/server"
)

func main() {
	templateFlag := flag.String("template", "vantage", "Template to use (vantage, basic, modern)")
	listFlag := flag.Bool("list", false, "List available templates")
	serveFlag := flag.Bool("serve", false, "Start web server")
	portFlag := flag.String("port", "8080", "Port to serve on")
	flag.Parse()

	// Initialize configuration and generator
	cfg := config.NewConfig()
	gen := generator.New(cfg)

	if *serveFlag {
		srv := server.New(gen)
		srv.SetupRoutes()
		fmt.Printf("Starting server on http://localhost:%s\n", *portFlag)

		httpServer := &http.Server{
			Addr:         ":" + *portFlag,
			Handler:      nil,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		log.Fatal(httpServer.ListenAndServe())
		return
	}

	if *listFlag {
		gen.ListTemplates()
		return
	}

	if err := gen.Generate(context.Background(), *templateFlag); err != nil {
		log.Fatalf("Error generating CV: %v", err)
	}
}
