package main

import (
	"context"
	"log"
	"net/http"

	"github.com/AlexTLDR/mycv.quest/components/layout"
)

func main() {
	// Set up the root route
	http.HandleFunc("/", homeHandler)

	// Serve static assets
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	log.Println("Server starting on port 3132...")
	log.Println("Visit http://localhost:3132")

	if err := http.ListenAndServe(":3132", nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Set content type to HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Write the HTML document structure
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>mycv.quest - Professional CV Generator</title>
    <link rel="stylesheet" href="/assets/css/output.css">
</head>
<body class="antialiased">
`))

	// Render the Layout003 component
	ctx := context.Background()
	if err := layout.Layout003().Render(ctx, w); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Close the HTML document
	w.Write([]byte(`
</body>
</html>`))
}
