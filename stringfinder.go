package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ASCII art for "Winfinder"
var winfinderLogo = []string{
	`W`,
	`i`,
	`n`,
	`f`,
	`i`,
	`n`,
	`d`,
	`e`,
	`r`,
}

var found bool = false

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

func printIntro() {
	for _, char := range winfinderLogo {
		fmt.Print(char)
		time.Sleep(150 * time.Millisecond) // Adjust the delay here to control the speed
	}
	fmt.Println("\n") // Move to the next line after the intro
}

func main() {
	printIntro() // Display the ASCII art intro
	// Handle no parameters case
	if len(os.Args) == 1 {
		fmt.Println("run with -help parameter to get commandline options")
	}

	// Check if the -help parameter is present
	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			if arg == "-help" {
				printHelp()
				return
			}
		}
	}

	var directory, searchString, fileExtension, regexPattern string
	var searchDepth int

	// Get directory, search string, file extension, search depth, and regex pattern from arguments or stdin
	if len(os.Args) < 6 {
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

		fmt.Print("Enter regex pattern (leave empty if not using regex): ")
		regexPattern, _ = reader.ReadString('\n')
		regexPattern = strings.TrimSpace(regexPattern)
	} else {
		directory = os.Args[1]
		searchString = os.Args[2]
		fileExtension = os.Args[3]
		searchDepth, _ = strconv.Atoi(os.Args[4])
		regexPattern = os.Args[5]
	}

	// Exit program if search depth is negative
	if searchDepth < 0 {
		fmt.Println("Error: search depth cannot be a negative number.")
		os.Exit(1)
	}

	if searchDepth == 0 {
		searchDepth = 1
	}

	// Initialize channels and variables for searching and displaying progress
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
		err := searchDirectory(directory, searchString, fileExtension, searchDepth, 0, foundFiles, progress, &totalFiles, regexPattern)
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
	found := false
	fmt.Println("\nSearch Results:")
	for filePath := range foundFiles {
		fmt.Printf("\033[32mFound in file: %s\033[0m\n", filePath)
		found = true
	}

	if !found {
		fmt.Println("\033[31mNo matching file found!\033[0m")
	}
}

func displayProgressBar(progress <-chan float64, totalFiles *int32) {
	for p := range progress {
		fmt.Printf("\rProgress: %.2f%%", math.Round(p*100*100)/100)
	}
	fmt.Printf("\rProgress: 100%% - %d files processed.\n", *totalFiles)
}

func searchDirectory(directory, searchString, fileExtension string, searchDepth, currentDepth int, foundFiles chan<- string, progress chan<- float64, totalFiles *int32, regexPattern string) error {
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
			err := searchDirectory(path, searchString, fileExtension, searchDepth, currentDepth+1, foundFiles, progress, totalFiles, regexPattern)
			if err != nil {
				return err
			}
		} else if strings.HasSuffix(file.Name(), fileExtension) {
			err := searchFile(path, searchString, foundFiles, regexPattern)
			if err != nil {
				return err
			}
			atomic.AddInt32(&processedFiles, 1)
			progress <- float64(atomic.LoadInt32(&processedFiles)) / float64(*totalFiles)
		}
	}

	return nil
}

func searchFile(filePath, searchString string, foundFiles chan<- string, regexPattern string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Compile the regex pattern if provided
	var regex *regexp.Regexp
	if regexPattern != "" {
		regex, err = regexp.Compile(regexPattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %v", err)
		}
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		// Use regex for searching if pattern is provided
		if regex != nil {
			if regex.MatchString(text) {
				foundFiles <- filePath
				found = true
				break
			}
		} else {
			// Fallback to simple string search if no regex pattern is provided
			if strings.Contains(text, searchString) {
				foundFiles <- filePath
				found = true
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func printHelp() {
	fmt.Println("Usage of this program:")
	fmt.Println("  [directory] [searchString] [fileExtension] [searchDepth] [regexPattern]")
	fmt.Println("Parameters:")
	fmt.Println("  directory: The directory to start the search in. Use '.' for current directory.")
	fmt.Println("  searchString: The string to search within files. Ignored if regexPattern is provided.")
	fmt.Println("  fileExtension: The extension of files to search in. Example: .txt")
	fmt.Println("  searchDepth: The depth of subdirectories to search in.")
	fmt.Println("  regexPattern: The regular expression pattern to search within files. Optional.")
	fmt.Println("\nExample:")
	fmt.Println("  ./program /path/to/directory \"search term\" .txt 2 \"^.*pr.*$\"")
	fmt.Println("\nUse -help to display this help message.")
}

func displaySearchResults(foundFiles <-chan string) {
	for filePath := range foundFiles {
		fmt.Printf("\033[32mFound in file: %s\033[0m\n", filePath)
	}
}
