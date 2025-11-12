package httpserver

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ecpartan/soap-server-tr069/internal/apperror"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/julienschmidt/httprouter"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) error {

	logger.LogDebug("Enter")
	filename := filepath.Base(r.URL.Path)
	if filename == "" || filename == "/" {
		return fmt.Errorf("Filename not provided in URL path")
	}

	logger.LogDebug("Enter", filename)

	filePath := filepath.Join("./uploads", filename)
	outFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	logger.LogDebug("Enter", filePath)
	_, err = io.Copy(outFile, r.Body)
	if err != nil {
		return err
	}

	logger.LogDebug("File  uploaded successfully to ", filename, filePath)
	return nil
}

func downloadHandler(w http.ResponseWriter, r *http.Request) error {

	logger.LogDebug("Enter", r.PathValue("id"))

	filePath := filepath.Join("./uploads", r.PathValue("id"))

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	fileName := filepath.Base(filePath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	_, err = io.Copy(w, file)
	if err != nil {
		return err
	}
	return nil
}

func NewFileServer(router *httprouter.Router) error {

	err := os.Mkdir("./uploads", 0755)
	if err != nil && !os.IsExist(err) {
		logger.LogDebug("Error creating directory:", err)
		return err
	}

	fs := http.FileServer(http.Dir("./uploads"))
	if fs == nil {
		logger.LogDebug("Error creating file server")
		return err
	}

	router.HandlerFunc(http.MethodPut, "/files/", apperror.Middleware(uploadHandler))
	router.HandlerFunc(http.MethodGet, "/files/{id}", apperror.Middleware(downloadHandler))

	return nil
}
