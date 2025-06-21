package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"YAMATO/dat"
	"YAMATO/ma0"
	"YAMATO/usage"
)

func uploadDatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["datFileInput[]"]
	if len(files) == 0 {
		http.Error(w, "No DAT file uploaded", http.StatusBadRequest)
		return
	}

	var allRecords []dat.DATRecord
	totalCount, ma0CreatedCount, duplicateCount := 0, 0, 0
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Println("Error opening DAT file:", err)
			continue
		}
		defer file.Close()

		records, tc, mc, dc, err := dat.ParseDATFile(file)
		if err != nil {
			log.Println("Error parsing DAT file:", err)
			continue
		}
		totalCount += tc
		ma0CreatedCount += mc
		duplicateCount += dc
		allRecords = append(allRecords, records...)
	}

	resp := map[string]interface{}{
		"DATReadCount":    totalCount,
		"MA0CreatedCount": ma0CreatedCount,
		"DuplicateCount":  duplicateCount,
		"DATRecords":      allRecords, // ここにDATの内容をすべて含む
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func uploadUsageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["usageFileInput[]"]
	if len(files) == 0 {
		http.Error(w, "No USAGE file uploaded", http.StatusBadRequest)
		return
	}
	var allRecords []usage.UsageRecord
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Println("Error opening USAGE file:", err)
			continue
		}
		defer file.Close()
		records, err := usage.ParseUsageFile(file)
		if err != nil {
			log.Println("Error parsing USAGE file:", err)
			continue
		}
		allRecords = append(allRecords, records...)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"USAGERecords": allRecords,
		"TotalRecords": len(allRecords),
	})
}

func viewMA0Handler(w http.ResponseWriter, r *http.Request) {
	ma0.ViewMA0Handler(w, r)
}

func autoLaunchBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	if err := exec.Command(cmd, args...).Start(); err != nil {
		log.Printf("Browser auto-launch failed: %v", err)
	}
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/uploadDat", uploadDatHandler)
	http.HandleFunc("/uploadUsage", uploadUsageHandler)
	http.HandleFunc("/viewMA0", viewMA0Handler)
	go autoLaunchBrowser("http://localhost:8080")
	log.Println("Server starting on port :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
