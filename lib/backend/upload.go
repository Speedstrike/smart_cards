package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/kidskoding/smart_cards/lib/backend/models"
)

func saveUploadedFile(file io.Reader, originalFilename string, size int64) (models.FileInfo, error) {
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return models.FileInfo{}, fmt.Errorf("failed to create upload directory: %v", err)
	}
	
	ext := filepath.Ext(originalFilename)
	uniqueFilename := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), uuid.New().String()[:8], ext)
	filePath := filepath.Join(uploadDir, uniqueFilename)
	
	outFile, err := os.Create(filePath)
	if err != nil {
		return models.FileInfo{}, fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()
	
	bytesWritten, err := io.Copy(outFile, file)
	if err != nil {
		return models.FileInfo{}, fmt.Errorf("failed to save file: %v", err)
	}
	
	contentType := getContentTypeFromExtension(ext)
	
	fileInfo := models.FileInfo{
		OriginalName: originalFilename,
		SavedPath:    filePath,
		Size:         bytesWritten,
		ContentType:  contentType,
	}
	
	return fileInfo, nil
}

func getContentTypeFromExtension(ext string) string {
	switch ext {
		case ".jpg", ".jpeg":
			return "image/jpeg"
		case ".png":
			return "image/png"
		case ".gif":
			return "image/gif"
		case ".pdf":
			return "application/pdf"
		case ".txt":
			return "text/plain"
		case ".doc":
			return "application/msword"
		case ".docx":
			return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		default:
			return "application/octet-stream"
	}
}

func generateSampleFlashcards(fileInfo models.FileInfo) []models.Flashcard {
	return []models.Flashcard{
		{
			ID:         uuid.New().String(),
			Question:   fmt.Sprintf("what was uploaded in the file %s?", fileInfo.OriginalName),
			Answer:     fmt.Sprintf("a %s file was uploaded with size %d bytes", fileInfo.ContentType, fileInfo.Size),
			Difficulty: "easy",
			Source:     fileInfo.OriginalName,
			CreatedAt:  time.Now().Format(time.RFC3339),
		},
	}
}

func uploadFilesHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("received upload request from %s", r.RemoteAddr)
	
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Printf("error parsing multipart form: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Error:   "failed to parse form",
			Details: err.Error(),
		})
		return
	}
	
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		log.Println("no files found in request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "No files uploaded"})
		return
	}
	
	log.Printf("processing %d files", len(files))
	
	var processedFiles []models.FileInfo
	var allFlashcards []models.Flashcard
	var errors []string
	
	for i, fileHeader := range files {
		log.Printf("processing file %d: %s (size: %d bytes)", i+1, fileHeader.Filename, fileHeader.Size)
		
		if fileHeader.Size > 10<<20 {
			errorMsg := fmt.Sprintf("file %s is too large (max 10MB)", fileHeader.Filename)
			log.Println(errorMsg)
			errors = append(errors, errorMsg)
			continue
		}
		
		ext := filepath.Ext(fileHeader.Filename)
		allowedExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".txt", ".doc", ".docx"}
		isAllowed := slices.Contains(allowedExtensions, ext)
		
		if !isAllowed {
			errorMsg := fmt.Sprintf("file type %s not allowed for %s", ext, fileHeader.Filename)
			log.Println(errorMsg)
			errors = append(errors, errorMsg)
			continue
		}
		
		file, err := fileHeader.Open()
		if err != nil {
			errorMsg := fmt.Sprintf("error opening file %s: %v", fileHeader.Filename, err)
			log.Println(errorMsg)
			errors = append(errors, errorMsg)
			continue
		}
		
		fileInfo, err := saveUploadedFile(file, fileHeader.Filename, fileHeader.Size)
		file.Close()
		
		if err != nil {
			errorMsg := fmt.Sprintf("error saving file %s: %v", fileHeader.Filename, err)
			log.Println(errorMsg)
			errors = append(errors, errorMsg)
			continue
		}
		
		log.Printf("successfully saved file: %s -> %s", fileInfo.OriginalName, fileInfo.SavedPath)
		processedFiles = append(processedFiles, fileInfo)
		
		flashcards := generateSampleFlashcards(fileInfo)
		allFlashcards = append(allFlashcards, flashcards...)
	}

	for _, fc := range allFlashcards {
		addToDB(fc)
	}
	
	if len(processedFiles) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Error:   "no files could be processed",
			Details: fmt.Sprintf("errors: %v", errors),
		})
		return
	}
	
	response := models.UploadResponse{
		Success:    true,
		Message:    fmt.Sprintf("successfully processed %d files", len(processedFiles)),
		Files:      processedFiles,
		Flashcards: allFlashcards,
	}
	
	log.Printf("upload completed successfully; processed %d files and generated %d flashcards", 
		len(processedFiles), len(allFlashcards))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func addToDB(flashcard models.Flashcard) {
	ctx := context.Background()
	rdb := connect()
	defer rdb.Close()

    err := rdb.HSet(ctx, "0", flashcard.Question, flashcard.Answer).Err()
    if err != nil {
        log.Fatal(err)
		return
    }
}
