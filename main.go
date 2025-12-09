package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
	"encoding/json"
)

const (
	unoserverHost = "127.0.0.1"
	unoserverPort = "2003"
	tempDir       = "/tmp/conversions"
	maxFileSize   = 50 << 20 // 50 MB
)

func init() {
	os.MkdirAll(tempDir, 0755)
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// http.Error(w, "Somente o método POST é aceito.", http.StatusMethodNotAllowed)
		sendError(w, "Somente o método POST é aceito.", http.StatusMethodNotAllowed, "")
		return
	}

	// Requer que o cliente envie a extensão via cabeçalho
	ext := r.Header.Get("X-File-Extension")
	if ext == "" {
		ext = "docx"
		// http.Error(w, "O header X-File-Extension é obrigatório.", http.StatusBadRequest)
		// return
	}

	// Ler corpo bruto (binário)
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)
	inputBytes, err := io.ReadAll(r.Body)
	if err != nil {
		// http.Error(w, "Erro ao ler o arquivo de entrada.", http.StatusBadRequest)
		sendError(w, "Erro ao ler o arquivo de entrada.", http.StatusBadRequest, err.Error())
		return
	}

	// Criar arquivos temporários
	inputFile, err := os.CreateTemp(tempDir, "in_*." + ext)
	if err != nil {
		// http.Error(w, "Erro ao criar o arquivo temporário de entrada.", http.StatusInternalServerError)
		sendError(w, "Erro ao criar o arquivo temporário de entrada.", http.StatusInternalServerError, err.Error())
		return
	}
	defer os.Remove(inputFile.Name())

	outputFile, err := os.CreateTemp(tempDir, "out_*.pdf")
	if err != nil {
		// http.Error(w, "Erro ao criar o arquivo de saída.", http.StatusInternalServerError)
		sendError(w, "Erro ao criar o arquivo de saída.", http.StatusInternalServerError, err.Error())
		return
	}
	defer os.Remove(outputFile.Name())

	// Escrever arquivo de entrada
	if _, err := inputFile.Write(inputBytes); err != nil {
		// http.Error(w, "Erro ao escrever no arquivo.", http.StatusInternalServerError)
		sendError(w, "Erro ao escrever no arquivo.", http.StatusInternalServerError, err.Error())
		return
	}
	inputFile.Close()

	// Executar conversão com timeout
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "unoconvert",
		"--host", unoserverHost,
		"--port", unoserverPort,
		"--host-location", "local",
		inputFile.Name(),
		outputFile.Name())

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("Falha na conversão da extensão '%s': %v | stderr: %s | stdout: %s", ext, err, stderr.String(), stdout.String())
		//http.Error(w, "Falha na conversão.", http.StatusInternalServerError)
		sendError(w, "Falha na conversão.", http.StatusInternalServerError, err.Error())
		return
	}

	// Ler e retornar PDF
	pdfBytes, err := os.ReadFile(outputFile.Name())
	if err != nil {
		// http.Error(w, "Falha ao ler o PDF criado.", http.StatusInternalServerError)
		sendError(w, "Falha ao ler o PDF criado.", http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))
	w.Write(pdfBytes)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	http.HandleFunc("/convert", convertHandler)
	http.HandleFunc("/health", healthHandler)
	log.Println("Conversor para PDF do LibreOffice escutando a porta 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func sendError(w http.ResponseWriter, message string, statusCode int, details string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	resp := ErrorResponse{
		Code:    statusCode,
		Message: message,
		Details: details,
	}
	json.NewEncoder(w).Encode(resp)
}