package handlers

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"ahavault/server/internal/middleware"
	"ahavault/server/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tus/tusd/v2/pkg/filestore"
	tusd "github.com/tus/tusd/v2/pkg/handler"
)

type TusHandler struct {
	Handler     *tusd.Handler
	fileService *services.FileService
	basePath    string
	uploadDir   string
}

func NewTusHandler(fileService *services.FileService, uploadDir string) *TusHandler {
	// Create upload directory if not exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("Failed to create tus upload directory: %v", err)
	}

	store := filestore.New(uploadDir)
	composer := tusd.NewStoreComposer()
	store.UseIn(composer)

	// Base path must match the router path prefix (without wildcard)
	basePath := "/api/tus/upload/"

	handler, err := tusd.NewHandler(tusd.Config{
		BasePath:              basePath,
		StoreComposer:         composer,
		NotifyCompleteUploads: true,
	})
	if err != nil {
		log.Fatalf("Unable to create tus handler: %s", err)
	}

	th := &TusHandler{
		Handler:     handler,
		fileService: fileService,
		basePath:    basePath,
		uploadDir:   uploadDir,
	}

	// Start background listener for completed uploads
	go th.handleCompletedUploads()

	return th
}

// GinHandler wraps the tusd handler for Gin
func (h *TusHandler) GinHandler(c *gin.Context) {
	log.Printf("TusHandler processing: %s %s", c.Request.Method, c.Request.URL.Path)
	
	// Strip the prefix to match what tusd expects if needed,
	// but tusd uses http.StripPrefix usually. 
	// In Gin, we can just pass the request. 
	// Important: Handle Metadata injection for POST (Creation)
	
	if c.Request.Method == "POST" {
		userID := middleware.GetUserID(c)
		if userID != "" {
			// Inject userID into Upload-Metadata
			// Format: key valueBase64,key2 value2Base64
			meta := c.Request.Header.Get("Upload-Metadata")
			
			// Encode userID to base64
			encodedInfo := base64.StdEncoding.EncodeToString([]byte(userID))
			kv := fmt.Sprintf("userID %s", encodedInfo)

			if meta == "" {
				meta = kv
			} else {
				meta = meta + "," + kv
			}
			c.Request.Header.Set("Upload-Metadata", meta)
		}
	}

	// ServeHTTP handles the request
	// Note: Gin's wildcard route might include the prefix in URL.Path
	// tusd expects the path *after* BasePath if we don't use StripPrefix.
	// But we configured BasePath in tusd config. 
	// Let's verify: URL /api/tus/upload/123 -> BasePath /api/tus/upload/ -> ID 123.
	// This should work directly.
	h.Handler.ServeHTTP(c.Writer, c.Request)
}

func (h *TusHandler) handleCompletedUploads() {
	for {
		event := <-h.Handler.CompleteUploads
		log.Printf("Tus upload %s finished", event.Upload.ID)

		// Fix: Extract Upload struct from HookEvent
		go h.processUpload(event.Upload)
	}
}

func (h *TusHandler) processUpload(upload tusd.FileInfo) {
	// Recover metadata
	// Fix: Access fields directly on upload struct (no .Upload nesting)
	meta := upload.MetaData
	userIDStr := meta["userID"]
	filename := meta["filename"]
	
	if filename == "" {
		filename = "uploaded_file"
	}
	if userIDStr == "" {
		log.Printf("Error: No userID in metadata for upload %s", upload.ID)
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Error processing upload %s: invalid user ID %s", upload.ID, userIDStr)
		return
	}

	// Open the uploaded file from disk
	filePath := fmt.Sprintf("%s/%s", h.uploadDir, upload.ID)
	// infoPath := filePath + ".info" (tusd creates .info files too)

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file %s: %v", filePath, err)
		return
	}
	defer file.Close()

	// Call FileService to process (encrypt and store permanently)
	// Note: Size needs to be int64
	_, err = h.fileService.UploadFile(userUUID, filename, upload.Size, file)
	if err != nil {
		log.Printf("Error saving file to permanent storage: %v", err)
		return
	}

	// Cleanup temp files (optional, or keep generic cleanup task)
	// tusd doesn't automatically delete completed files from store?
	// We should remove it after successful processing.
	file.Close() // Ensure closed before removal
	if err := os.Remove(filePath); err != nil {
		log.Printf("Warning: failed to remove temp file %s: %v", filePath, err)
	}
	if err := os.Remove(filePath + ".info"); err != nil {
		log.Printf("Warning: failed to remove info file: %v", err)
	}
	
	log.Printf("Successfully processed and stored upload %s", upload.ID)
}
