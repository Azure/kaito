package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

const ExternalIP = "20.4.240.121"
const ExternalPort = "80"
const BaseURL = "http://" + ExternalIP + ":" + ExternalPort + "/download/"

func downloadFile(fp string, url string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Create the file
	out, err := os.Create(fp)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Track progress
	totalBytes := resp.ContentLength
	var bytesRead int64

	progressReader := io.TeeReader(resp.Body, &WriteCounter{filename: fp, total: totalBytes, read: &bytesRead})

	// Write the data to the file
	_, err = io.Copy(out, progressReader)
	if err != nil {
		log.Fatal(err)
	}
}

type WriteCounter struct {
	filename     string
	total        int64
	read         *int64
	lastReported int64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	*wc.read += int64(n)

	// Calculate every 1% increment of the total size
	onePercent := wc.total / 100

	// Check if the bytes read has surpassed another 1% increment since the last reported value
	if *wc.read-wc.lastReported >= onePercent {
		fmt.Printf("\rDownloading [%s]: %d out of %d bytes (%.2f%%)\n", filepath.Base(wc.filename), *wc.read, wc.total, float64(*wc.read)/float64(wc.total)*100)
		wc.lastReported = *wc.read
	}

	return n, nil
}

func getURLsForModel(modelVersion string) []string {
	switch modelVersion {
	case "llama-2-7b":
		return []string{
			BaseURL + "llama-2-7b/consolidated.00.pth",
		}
	case "llama-2-7b-chat":
		return []string{
			BaseURL + "llama-2-7b-chat/consolidated.00.pth",
		}
	case "llama-2-13b":
		return []string{
			BaseURL + "llama-2-13b/consolidated.00.pth",
			BaseURL + "llama-2-13b/consolidated.01.pth",
		}
	case "llama-2-13b-chat":
		return []string{
			BaseURL + "llama-2-13b-chat/consolidated.00.pth",
			BaseURL + "llama-2-13b-chat/consolidated.01.pth",
		}
	case "llama-2-70b":
		return []string{
			BaseURL + "llama-2-70b/consolidated.00.pth",
			BaseURL + "llama-2-70b/consolidated.01.pth",
			BaseURL + "llama-2-70b/consolidated.02.pth",
			BaseURL + "llama-2-70b/consolidated.03.pth",
			BaseURL + "llama-2-70b/consolidated.04.pth",
			BaseURL + "llama-2-70b/consolidated.05.pth",
			BaseURL + "llama-2-70b/consolidated.06.pth",
			BaseURL + "llama-2-70b/consolidated.07.pth",
		}
	case "llama-2-70b-chat":
		return []string{
			BaseURL + "llama-2-70b-chat/consolidated.00.pth",
			BaseURL + "llama-2-70b-chat/consolidated.01.pth",
			BaseURL + "llama-2-70b-chat/consolidated.02.pth",
			BaseURL + "llama-2-70b-chat/consolidated.03.pth",
			BaseURL + "llama-2-70b-chat/consolidated.04.pth",
			BaseURL + "llama-2-70b-chat/consolidated.05.pth",
			BaseURL + "llama-2-70b-chat/consolidated.06.pth",
			BaseURL + "llama-2-70b-chat/consolidated.07.pth",
		}
	default:
		log.Fatalf("Invalid model version: %s", modelVersion)
		return nil
	}
}

func ensureDirExists(dirName string) {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := os.Mkdir(dirName, 0755)
		if err != nil {
			log.Fatalf("Failed to create directory: %v", err)
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <model_version>", os.Args[0])
	}

	ensureDirExists("weights")

	modelVersion := os.Args[1]
	urls := getURLsForModel(modelVersion)

	var wg sync.WaitGroup

	for i, url := range urls {
		fp := fmt.Sprintf("weights/consolidated.%02d.pth", i)
		wg.Add(1)
		go downloadFile(fp, url, &wg)
	}

	wg.Wait()
}
