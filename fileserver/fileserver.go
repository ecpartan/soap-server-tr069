package fileserver

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ecpartan/soap-server-tr069/internal/apperror"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/pkg/httplogger"
	"github.com/julienschmidt/httprouter"
)

/*
func UploadHandler(w http.ResponseWriter, r *http.Request) error {

	logger.LogDebug("Enter to uploadHandler")
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

func DownloadHandler(w http.ResponseWriter, r *http.Request) error {

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
*/

const (
	uploadDir   = "./uploads" // директория для хранения файлов
	maxFileSize = 10 << 20    // 10 MB
)

func DownloadHandler(w http.ResponseWriter, r *http.Request) error {

	params := httprouter.ParamsFromContext(r.Context())

	logger.LogDebug("Enter DownloadHandler", params.ByName("id"))

	filePath := filepath.Join(uploadDir, params.ByName("id"))
	logger.LogDebug("Enter DownloadHandler", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		logger.LogDebug("Enter DownloadHandler", err)

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

func UploadHandler(w http.ResponseWriter, r *http.Request) error {
	logger.LogDebug("Enter to uploadHandler")
	filename := filepath.Base(r.URL.Path)
	if filename == "" || filename == "/" {
		return fmt.Errorf("Filename not provided in URL path")
	}

	logger.LogDebug("Enter", filename)

	filePath := filepath.Join(uploadDir, filename)
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

// formatSize - форматирует размер файла в человеко-читаемый вид
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprint(bytes) + " B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprint(bytes/div) + " " + string("KMGTPE"[exp]) + "B"
}
func ListHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return fmt.Errorf("Method not allowed %w", http.StatusMethodNotAllowed)

	}

	files, err := os.ReadDir(uploadDir)
	if err != nil {
		return fmt.Errorf("Don't reading dir %w", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<h1>File list:</h1><ul>"))

	for _, file := range files {
		if !file.IsDir() {
			filename := file.Name()
			info, _ := file.Info()
			size := info.Size()
			w.Write([]byte("<li>"))
			w.Write([]byte(filename + " (" + formatSize(size) + ") - "))
			w.Write([]byte("<a href='/download/" + filename + "'>Download</a> | "))
			w.Write([]byte("<a href='/delete/" + filename + "'>Delete</a>"))
			w.Write([]byte("</li>"))
		}
	}
	w.Write([]byte("</ul>"))
	w.Write([]byte("<h2>Download file:</h2>"))
	w.Write([]byte(`<form method="post" action="/upload" enctype="multipart/form-data">
		<input type="file" name="file" required>
		<input type="submit" value="Download">
	</form>`))
	return nil
}

func NewFileServer(router *httprouter.Router) error {

	logger.LogDebug("Enter to create fileserver")
	err := os.Mkdir(uploadDir, 0755)
	if err != nil && !os.IsExist(err) {
		logger.LogDebug("Error creating directory:", err)
		return err
	}
	middleware := httplogger.NewHTTPLogger().Middleware
	router.Handler(http.MethodPut, "/upload/", httplogger.NewHTTPLogger().Middleware(apperror.Middleware(UploadHandler)))
	router.Handler(http.MethodGet, "/upload/:id", middleware(apperror.Middleware(DownloadHandler)))
	router.Handler(http.MethodGet, "/files/", middleware(apperror.Middleware(ListHandler)))

	return nil
}
