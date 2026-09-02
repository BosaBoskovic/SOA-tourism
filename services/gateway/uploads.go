package main

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gateway/auth"
)

// Media upload: replaces the base64-inline-JSON pattern used for
// profile/keypoint/review/blog images. A file goes here once, comes back as
// a short URL, and every other service keeps treating "image" as the plain
// string field it already is - no schema changes needed anywhere else.
const maxUploadSize = 5 << 20 // 5MB

var allowedUploadExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

func uploadsDir() string {
	dir := getEnvOrDefault("UPLOADS_DIR", "/data/uploads")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

// uploadHandler accepts a single multipart "file" field and returns
// {"url": "/uploads/<name>"} - callers store that URL wherever they
// currently store an image string (profile.imageURL, keypoint.imageUrl, ...).
// Unlike the rest of the gateway (which trusts the backend behind each
// proxied route to verify the JWT it forwards), this endpoint is served by
// the gateway itself, so it verifies the token directly.
func uploadHandler(verifier *auth.Verifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "Metod nije dozvoljen"})
			return
		}
		if _, err := verifier.Parse(r.Header.Get("Authorization")); err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "Autentifikacija je obavezna"})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+1<<20) // small margin for multipart overhead
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Fajl je prevelik ili zahtev nije ispravan (max 5MB)"})
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Nedostaje 'file' polje"})
			return
		}
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(header.Filename))
		if !allowedUploadExtensions[ext] {
			writeJSON(w, http.StatusUnsupportedMediaType, map[string]any{"error": "Dozvoljeni formati: jpg, jpeg, png, gif, webp"})
			return
		}

		name, err := randomFilename(ext)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "Greska pri generisanju imena fajla"})
			return
		}

		dst, err := os.Create(filepath.Join(uploadsDir(), name))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "Greska pri cuvanju fajla"})
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "Greska pri cuvanju fajla"})
			return
		}

		writeJSON(w, http.StatusCreated, map[string]any{"url": "/uploads/" + name})
	}
}

func randomFilename(ext string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf) + ext, nil
}
