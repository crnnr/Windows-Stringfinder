package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

func countFiles(directory, fileExtension string, searchDepth, currentDepth int) (int32, error) {
	if currentDepth > searchDepth {
		return 0, nil
	}

	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return 0, err
	}

	count := int32(0)
	for _, file := range files {
		path := filepath.Join(directory, file.Name())

		if file.IsDir() {
			subDirCount, err := countFiles(path, fileExtension, searchDepth, currentDepth+1)
			if err != nil {
				return 0, err
			}
			count += subDirCount
		} else if strings.HasSuffix(file.Name(), fileExtension) {
			count++
		}
	}

	return count, nil
}

func main() {
	var directory, searchString, fileExtension string
	var searchDepth int

	// Get directory, search string, file extension, and search depth from arguments or stdin
	if len(os.Args) < 5 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter directory (leave empty for current directory): ")
		directory, _ = reader.ReadString('\n')
		directory = strings.TrimSpace(directory)

		if directory == "" {
			directory = "."
		}

		fmt.Print("Enter string to search: ")
		searchString, _ = reader.ReadString('\n')
		searchString = strings.TrimSpace(searchString)

		fmt.Print("Enter file extension to search (.txt): ")
		fileExtension, _ = reader.ReadString('\n')
		fileExtension = strings.TrimSpace(fileExtension)

		fmt.Print("Enter search depth (default is 1): ")
		fmt.Fscanln(reader, &searchDepth)
		if searchDepth == 0 {
			searchDepth = 1
		}
	} else {
		directory = os.Args[1]
		searchString = os.Args[2]
		fileExtension = os.Args[3]
		searchDepth, _ = strconv.Atoi(os.Args[4])
		if searchDepth == 0 {
			searchDepth = 1
		}
	}

	// Create a channel to signal the completion of the search and for progress updates
	progress := make(chan float64)
	foundFiles := make(chan string, 100) // buffer size can be adjusted

	var totalFiles int32
	totalFiles, _ = countFiles(directory, fileExtension, searchDepth, 0)
	var wg sync.WaitGroup
	wg.Add(1) // Only adding 1 because only the searchDirectory goroutine signals the WaitGroup

	// Start the loading animation in a separate goroutine
	go displayProgressBar(progress, &totalFiles)

	// Perform the search
	go func() {
		defer wg.Done()
		err := searchDirectory(directory, searchString, fileExtension, searchDepth, 0, foundFiles, progress, &totalFiles)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		close(foundFiles) // Close the channel to indicate no more files will be sent
		close(progress)   // Close the progress channel to signal completion
	}()

	// Wait for the search to complete
	wg.Wait()

	// Process and display found files
	fmt.Println("\nSearch Results:")
	displaySearchResults(foundFiles)
}

func displayProgressBar(progress <-chan float64, totalFiles *int32) {
	for p := range progress {
		fmt.Printf("\rProgress: %.2f%%", math.Round(p*100*100)/100)
	}
	fmt.Printf("\rProgress: 100%% - %d files processed.\n", *totalFiles)
}

func searchDirectory(directory, searchString, fileExtension string, searchDepth, currentDepth int, foundFiles chan<- string, progress chan<- float64, totalFiles *int32) error {
	if currentDepth > searchDepth {
		return nil
	}

	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}
	var processedFiles int32
	for _, file := range files {
		path := filepath.Join(directory, file.Name())

		if file.IsDir() {
			err := searchDirectory(path, searchString, fileExtension, searchDepth, currentDepth+1, foundFiles, progress, totalFiles)
			if err != nil {
				return err
			}
		} else if strings.HasSuffix(file.Name(), fileExtension) {
			atomic.AddInt32(totalFiles, 1)
			err := searchFile(path, searchString, foundFiles)
			if err != nil {
				return err
			}
			atomic.AddInt32(&processedFiles, 1)
			progress <- float64(atomic.LoadInt32(&processedFiles)) / float64(*totalFiles)
		}
	}

	return nil
}

func searchFile(filePath, searchString string, foundFiles chan<- string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	// Skip if it's a directory
	if fileInfo.IsDir() {
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	const chunkSize = 512 * 1024 // for example, 512KB
	buf := make([]byte, chunkSize)

	found := false
	for {
		bytesRead, err := file.Read(buf)
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		if strings.Contains(string(buf[:bytesRead]), searchString) {
			if !found {
				foundFiles <- filePath
				found = true
			}
		}
	}

	return nil
}

func displaySearchResults(foundFiles <-chan string) {
	for filePath := range foundFiles {
		fmt.Println("Found in file:", filePath)
	}
}
