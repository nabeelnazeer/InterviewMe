package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func UploadFile(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check file extension
	if filepath.Ext(handler.Filename) != ".pdf" {
		http.Error(w, "Only PDF files are allowed", http.StatusBadRequest)
		return
	}

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := os.CreateTemp("uploads", "upload-*.pdf")
	if err != nil {
		http.Error(w, "Unable to create the file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	// Copy the uploaded file to the temporary file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		http.Error(w, "Unable to save the file", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
}
