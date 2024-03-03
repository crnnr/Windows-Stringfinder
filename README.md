# Windows-Stringfinder

## Overview
Windows-Stringfinder is a powerful and efficient utility written in Go, designed to enhance your file search capabilities. It allows you to quickly find strings within files across multiple formats, offering a robust alternative to traditional search tools like `findstr`. With customizable search parameters, it's built for those who need a more targeted and versatile search utility.

## Features
- **Fast and Efficient**: Optimized to quickly search through large directories.
- **Support for Multiple File Formats**: Works with various file extensions, making it versatile for different use cases.
- **Customizable Search Parameters**: Allows fine-tuning of search depth, file extensions, and the use of regular expressions for advanced searches.

## Getting Started

### Prerequisites
- Ensure you have Go installed on your machine. You can download Go from [the official Go website](https://golang.org/dl/).

### Installation
1. Clone this repository to your local machine using Git:
   ```sh
   git clone https://github.com/your-username/Windows-Stringfinder.git
   ```
2. Navigate to the cloned repository:
   ```sh
   cd Windows-Stringfinder
   ```
3. Build the project (optional):
   ```sh
   go build
   ```

## Usage
After installation, you can start using Windows-Stringfinder to search for strings in files. Here's how to use it:

### Basic Usage
Run the program without any arguments to see the ASCII art intro and follow the interactive prompts:
```sh
./Windows-Stringfinder
```

### Advanced Usage
To directly specify the parameters without interactive prompts, use the following syntax:
```sh
./Windows-Stringfinder [directory] [searchString] [fileExtension] [searchDepth] [regexPattern]
```

#### Parameters:
- `directory`: The directory to start the search in. Use `.` for the current directory.
- `searchString`: The string to search within files. Ignored if `regexPattern` is provided.
- `fileExtension`: The extension of files to search in. Example: `.txt`.
- `searchDepth`: The depth of subdirectories to search in.
- `regexPattern`: The regular expression pattern to search within files. Optional.

#### Example:
Search for the term "example" in `.txt` files within the current directory and subdirectories up to a depth of 2:
```sh
./Windows-Stringfinder . "example" .txt 2
```

### Using Regular Expressions
For more advanced searches, you can use regular expressions. For example, to find files containing lines that start with "error":
```sh
./Windows-Stringfinder /path/to/directory "" .log 3 "^error"
```

### Help
Run the program with the `-help` parameter to display help information:
```sh
./Windows-Stringfinder -help
```

## Contributing
If you have improvements or bug fixes, just submit a pull request.