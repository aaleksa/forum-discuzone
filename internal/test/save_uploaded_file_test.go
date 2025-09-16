package test

import (
	"bytes"
	"forum/internal/utils"
	"mime/multipart"
	"testing"
)

func createTestFile(t *testing.T, filename, contentType string, content []byte) (multipart.File, *multipart.FileHeader) {
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	part, err := writer.CreateFormFile("image", filename)
	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write(content)
	if err != nil {
		t.Fatal(err)
	}
	writer.Close()

	// simulate file upload
	fileReader := bytes.NewReader(b.Bytes())
	req := multipart.NewReader(fileReader, writer.Boundary())

	form, err := req.ReadForm(25 << 20) // 25MB

	if err != nil {
		t.Fatal(err)
	}

	files := form.File["image"]
	if len(files) == 0 {
		t.Fatal("No file found in form")
	}

	opened, err := files[0].Open()
	if err != nil {
		t.Fatal(err)
	}

	return opened, files[0]
}

func TestSaveUploadedFile_LargeFile(t *testing.T) {
	large := make([]byte, 21<<20) // 21MB
	file, header := createTestFile(t, "big.jpg", "image/jpeg", large)
	defer file.Close()

	_, err := utils.SaveUploadedFile(file, header)
	if err == nil || err.Error() != "file is too large" {
		t.Fatalf("expected size error, got: %v", err)
	}
}

func TestSaveUploadedFile_UnsupportedType(t *testing.T) {
	content := []byte("%PDF-1.4") // Not an image
	file, header := createTestFile(t, "test.pdf", "application/pdf", content)
	defer file.Close()

	_, err := utils.SaveUploadedFile(file, header)
	if err == nil || err.Error() != "unsupported file type" {
		t.Fatalf("expected type error, got: %v", err)
	}
}
