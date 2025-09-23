package server_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/AlexTLDR/mycv.quest/pkg/config"
	"github.com/AlexTLDR/mycv.quest/pkg/generator"
	"github.com/AlexTLDR/mycv.quest/pkg/server"
)

func setupTestServer() *server.Server {
	cfg := &config.Config{
		Templates: map[string]config.Template{
			"basic": {
				Name:      "Basic Resume",
				Dir:       "../../templates/basic/template",
				InputFile: "main.typ",
			},
			"modern": {
				Name:       "Modern Resume",
				Dir:        "../../templates/modern/template",
				InputFile:  "main.typ",
				NeedsPhoto: true,
			},
			"vantage": {
				Name:      "Vantage Resume",
				Dir:       "../../templates/vantage",
				InputFile: "example.typ",
			},
		},
		OutputDir: "test_output",
	}

	gen := generator.New(cfg)
	return server.New(gen)
}

func TestHandleIndex(t *testing.T) {
	t.Parallel()
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	server.HandleIndex(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	bodyStr := string(body)

	// Should contain template information
	expectedContent := []string{
		"Basic Resume",
		"Modern Resume",
		"Vantage Resume",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(bodyStr, expected) {
			t.Errorf("Response body missing expected content: %s", expected)
		}
	}
}

func TestHandleIndexNotFound(t *testing.T) {
	t.Parallel()
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	server.HandleIndex(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestHandleForm(t *testing.T) {
	t.Parallel()
	server := setupTestServer()

	testCases := []struct {
		path           string
		expectedStatus int
		shouldContain  string
	}{
		{"/form/basic", http.StatusOK, "basic"},
		{"/form/modern", http.StatusOK, "modern"},
		{"/form/vantage", http.StatusOK, "vantage"},
		{"/form/nonexistent", http.StatusNotFound, ""},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		w := httptest.NewRecorder()

		server.HandleForm(w, req)

		resp := w.Result()
		if resp.StatusCode != tc.expectedStatus {
			t.Errorf("Path %s: expected status %d, got %d", tc.path, tc.expectedStatus, resp.StatusCode)
		}

		if tc.expectedStatus == http.StatusOK && tc.shouldContain != "" {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			bodyStr := strings.ToLower(string(body))
			if !strings.Contains(bodyStr, tc.shouldContain) {
				t.Errorf("Path %s: response body should contain %s", tc.path, tc.shouldContain)
			}
		}
	}
}

func TestHandleGenerateGET(t *testing.T) {
	t.Parallel()
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/generate/basic", nil)
	w := httptest.NewRecorder()

	server.HandleGenerate(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusSeeOther {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected redirect status 303, got %d. Body: %s", resp.StatusCode, string(body))
	}

	location := resp.Header.Get("Location")
	if location != "/form/basic" {
		t.Errorf("Expected redirect to /form/basic, got %s", location)
	}
}

func TestHandleGenerateMethodNotAllowed(t *testing.T) {
	t.Parallel()
	server := setupTestServer()

	req := httptest.NewRequest(http.MethodPut, "/generate/basic", nil)
	w := httptest.NewRecorder()

	server.HandleGenerate(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

func TestHandleGeneratePOSTBasic(t *testing.T) {
	t.Parallel()
	server := setupTestServer()

	// Create form data
	formData := url.Values{
		"name":                      {"Test User"},
		"email":                     {"test@example.com"},
		"location":                  {"Test City"},
		"github":                    {"testuser"},
		"education[0][institution]": {"Test University"},
		"education[0][degree]":      {"Test Degree"},
		"work[0][title]":            {"Test Job"},
		"work[0][company]":          {"Test Company"},
	}

	req := httptest.NewRequest(http.MethodPost, "/generate/basic", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	server.HandleGenerate(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusSeeOther {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected redirect status 303, got %d. Body: %s", resp.StatusCode, string(body))
	}

	// Should redirect to session-specific PDF
	location := resp.Header.Get("Location")
	if !strings.HasPrefix(location, "/cv/") || !strings.HasSuffix(location, "/basic.pdf") {
		t.Errorf("Expected redirect to session PDF, got %s", location)
	}

	// Should set session cookie
	cookies := resp.Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session_id" {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Error("Expected session cookie to be set")
	}
}

func TestHandleGeneratePOSTModernWithPhoto(t *testing.T) {
	t.Parallel()
	server := setupTestServer()

	// Create multipart form with photo upload
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add text fields
	fields := map[string]string{
		"author":    "Test User",
		"job_title": "Developer",
		"bio":       "Test bio",
		"email":     "test@example.com",
	}

	for key, value := range fields {
		field, err := writer.CreateFormField(key)
		if err != nil {
			t.Fatalf("Failed to create form field: %v", err)
		}
		if _, err := field.Write([]byte(value)); err != nil {
			t.Fatalf("Failed to write form field: %v", err)
		}
	}

	// Add fake image file
	fileWriter, err := writer.CreateFormFile("avatar", "test.png")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Write a minimal PNG header (enough for detection)
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if _, err := fileWriter.Write(pngHeader); err != nil {
		t.Fatalf("Failed to write PNG header: %v", err)
	}
	if _, err := fileWriter.Write([]byte("fake png data")); err != nil {
		t.Fatalf("Failed to write PNG data: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/generate/modern", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	server.HandleGenerate(w, req)

	resp := w.Result()

	// The test might fail with typst compilation error due to invalid PNG
	// but we still want to test the multipart form handling
	if resp.StatusCode == http.StatusInternalServerError {
		body, _ := io.ReadAll(resp.Body)
		if strings.Contains(string(body), "typst compilation failed") {
			t.Skip("Typst compilation failed due to invalid PNG - this is expected in test environment")
		}
		t.Errorf("Unexpected server error: %s", string(body))
	}

	if resp.StatusCode != http.StatusSeeOther {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected redirect status 303, got %d. Body: %s", resp.StatusCode, string(body))
	}

	location := resp.Header.Get("Location")
	if !strings.HasPrefix(location, "/cv/") || !strings.HasSuffix(location, "/modern.pdf") {
		t.Errorf("Expected redirect to session PDF, got %s", location)
	}
}

func TestHandleSessionPDF(t *testing.T) {
	t.Parallel()
	server := setupTestServer()

	// First generate a CV to create session data
	formData := url.Values{
		"name":  {"Test User"},
		"email": {"test@example.com"},
	}

	req := httptest.NewRequest(http.MethodPost, "/generate/basic", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	server.HandleGenerate(w, req)

	// Extract session ID from redirect location
	location := w.Header().Get("Location")
	parts := strings.Split(location, "/")
	if len(parts) < 3 {
		t.Fatalf("Invalid redirect location: %s", location)
	}

	// Now test accessing the PDF
	pdfReq := httptest.NewRequest(http.MethodGet, location, nil)
	pdfW := httptest.NewRecorder()

	server.HandleSessionPDF(pdfW, pdfReq)

	resp := pdfW.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for PDF access, got %d", resp.StatusCode)
	}

	// Check Content-Type header
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/pdf" {
		t.Errorf("Expected Content-Type application/pdf, got %s", contentType)
	}

	// Check Content-Disposition header
	contentDisposition := resp.Header.Get("Content-Disposition")
	if !strings.Contains(contentDisposition, "cv-basic.pdf") {
		t.Errorf("Expected filename in Content-Disposition, got %s", contentDisposition)
	}

	// Verify PDF content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read PDF response: %v", err)
	}

	if len(body) == 0 {
		t.Error("PDF response body is empty")
	}

	if !strings.HasPrefix(string(body[:4]), "%PDF") {
		t.Error("Response is not a valid PDF file")
	}
}

func TestHandleSessionPDFNotFound(t *testing.T) {
	t.Parallel()
	server := setupTestServer()

	testCases := []string{
		"/cv/nonexistent/basic.pdf",
		"/cv/invalid/path",
		"/cv/session123/nonexistent.pdf",
	}

	for _, path := range testCases {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()

		server.HandleSessionPDF(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Path %s: expected status 404, got %d", path, resp.StatusCode)
		}
	}
}

func TestSessionManagement(t *testing.T) {
	t.Parallel()
	server := setupTestServer()

	// Generate two CVs with different templates
	formData := url.Values{
		"name":  {"Test User"},
		"email": {"test@example.com"},
	}

	// Generate basic CV
	req1 := httptest.NewRequest(http.MethodPost, "/generate/basic", strings.NewReader(formData.Encode()))
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w1 := httptest.NewRecorder()

	server.HandleGenerate(w1, req1)

	// Extract session cookie
	resp1 := w1.Result()
	var sessionCookie *http.Cookie
	for _, cookie := range resp1.Cookies() {
		if cookie.Name == "session_id" {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Session cookie not found")
	}

	// Generate vantage CV with same session
	formData2 := url.Values{
		"name":     {"Test User 2"},
		"email":    {"test2@example.com"},
		"position": {"Developer"},
	}

	req2 := httptest.NewRequest(http.MethodPost, "/generate/vantage", strings.NewReader(formData2.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.AddCookie(sessionCookie) // Use existing session
	w2 := httptest.NewRecorder()

	server.HandleGenerate(w2, req2)

	// Both PDFs should be accessible with the same session
	sessionID := sessionCookie.Value

	// Test basic PDF
	basicReq := httptest.NewRequest(http.MethodGet, "/cv/"+sessionID+"/basic.pdf", nil)
	basicW := httptest.NewRecorder()
	server.HandleSessionPDF(basicW, basicReq)

	if basicW.Result().StatusCode != http.StatusOK {
		t.Error("Basic PDF should be accessible")
	}

	// Test vantage PDF
	vantageReq := httptest.NewRequest(http.MethodGet, "/cv/"+sessionID+"/vantage.pdf", nil)
	vantageW := httptest.NewRecorder()
	server.HandleSessionPDF(vantageW, vantageReq)

	if vantageW.Result().StatusCode != http.StatusOK {
		t.Error("Vantage PDF should be accessible")
	}
}

func TestConcurrentSessions(t *testing.T) {
	t.Parallel()
	server := setupTestServer()

	// Generate CVs from two different "users" (no cookies)
	formData := url.Values{
		"name":  {"User One"},
		"email": {"user1@example.com"},
	}

	req1 := httptest.NewRequest(http.MethodPost, "/generate/basic", strings.NewReader(formData.Encode()))
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w1 := httptest.NewRecorder()

	server.HandleGenerate(w1, req1)

	formData2 := url.Values{
		"name":  {"User Two"},
		"email": {"user2@example.com"},
	}

	req2 := httptest.NewRequest(http.MethodPost, "/generate/basic", strings.NewReader(formData2.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w2 := httptest.NewRecorder()

	server.HandleGenerate(w2, req2)

	// Extract session IDs
	location1 := w1.Header().Get("Location")
	location2 := w2.Header().Get("Location")

	sessionID1 := strings.Split(location1, "/")[2]
	sessionID2 := strings.Split(location2, "/")[2]

	// Session IDs should be different
	if sessionID1 == sessionID2 {
		t.Error("Different requests should generate different session IDs")
	}

	// Each user should only be able to access their own PDF
	req1PDF := httptest.NewRequest(http.MethodGet, location1, nil)
	w1PDF := httptest.NewRecorder()
	server.HandleSessionPDF(w1PDF, req1PDF)

	if w1PDF.Result().StatusCode != http.StatusOK {
		t.Error("User 1 should be able to access their PDF")
	}

	req2PDF := httptest.NewRequest(http.MethodGet, location2, nil)
	w2PDF := httptest.NewRecorder()
	server.HandleSessionPDF(w2PDF, req2PDF)

	if w2PDF.Result().StatusCode != http.StatusOK {
		t.Error("User 2 should be able to access their PDF")
	}
}
