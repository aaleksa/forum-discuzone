package utils

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func SaveUploadedFile(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	log.Printf("=== SaveUploadedFile called ===")
	log.Printf("Filename: %s", fileHeader.Filename)
	log.Printf("File size: %d", fileHeader.Size)

	// Validate file size (20MB limit)
	const maxSize = 20 * 1024 * 1024 // 20MB
	if fileHeader.Size > maxSize {
		log.Printf("ERROR: File too large: %d bytes (max: %d)", fileHeader.Size, maxSize)
		return "", fmt.Errorf("file is too large")
	}

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
	}

	// For content type detection, use the filename extension as fallback
	contentType := fileHeader.Header.Get("Content-Type")
	log.Printf("Content-Type from header: %s", contentType)

	// If no content type in header, detect from filename extension
	if contentType == "" {
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		switch ext {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		default:
			// Read first 512 bytes for detection
			buffer := make([]byte, 512)
			n, readErr := file.Read(buffer)
			if readErr != nil && readErr != io.EOF {
				log.Printf("ERROR: Failed to read file for type detection: %v", readErr)
				return "", fmt.Errorf("failed to read file")
			}
			contentType = http.DetectContentType(buffer[:n])

			// Reset file to beginning by reopening it
			file.Close()
			var reopenErr error
			file, reopenErr = fileHeader.Open()
			if reopenErr != nil {
				log.Printf("ERROR: Failed to reopen file: %v", reopenErr)
				return "", fmt.Errorf("failed to reopen file")
			}
		}
	}
	log.Printf("Detected content type: %s", contentType)

	if !allowedTypes[contentType] {
		log.Printf("ERROR: Unsupported file type: %s", contentType)
		return "", fmt.Errorf("unsupported file type")
	}

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" {
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		default:
			ext = ".jpg"
		}
	}

	// Generate unique filename using timestamp + random number
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	log.Printf("Generated filename: %s", filename)

	// Create uploads directory if it doesn't exist
	exe, _ := os.Executable()
	exeDir := filepath.Dir(exe)
	uploadsDir := filepath.Join(exeDir, "static", "uploads")
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Printf("ERROR: Failed to create uploads directory: %v", err)
		return "", fmt.Errorf("failed to create uploads directory")
	}

	// Full file path
	filePath := filepath.Join(uploadsDir, filename)
	log.Printf("Full file path: %s", filePath)

	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("ERROR: Failed to create file %s: %v", filePath, err)
		return "", fmt.Errorf("failed to create file")
	}
	defer dst.Close()

	// Write file content to disk
	log.Println("== START COPY ==")
	bytesWritten, err := io.Copy(dst, file)
	log.Println("== END COPY ==")
	if err != nil {
		log.Printf("ERROR: Failed to copy file content: %v", err)
		return "", fmt.Errorf("failed to copy file content")
	}

	log.Printf("Successfully saved file: %s (%d bytes written)", filePath, bytesWritten)

	// Return relative path for database storage
	relativePath := "/static/uploads/" + filename
	log.Printf("Returning relative path: %s", relativePath)

	return relativePath, nil
}
