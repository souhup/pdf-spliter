package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/spf13/cobra"
)

func main() {
	initFlags()
}

var rootCmd = &cobra.Command{
	Use:     "pdf-split",
	Short:   "PDF File Splitter by table of contents",
	Long:    `A command-line tool for splitting PDF files into multiple files according to the table of contents.`,
	RunE:    splitPDF,
	Example: `./pdf-split -i example.pdf -o output_dir`,
}

var (
	inputFilePath string
	outputDir     string
)

// initFlags initializes command line flags and validates required parameters.
// The program will terminate if required parameters are missing or parsing fails.
func initFlags() {
	rootCmd.Flags().StringVarP(&inputFilePath, "input", "i", "", "input file path")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "output", "output directory path")
	if err := rootCmd.MarkFlagRequired("input"); err != nil {
		log.Fatalf("failed to parse param: %v", err)
	}
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("failed to execute: %v", err)
	}
}

// splitPDF coordinates the PDF splitting process by reading bookmarks
// and creating separate files for each chapter.
// Parameters _ and _ are used to satisfy the cobra.Command RunE interface.
func splitPDF(_ *cobra.Command, _ []string) error {
	// Open the source PDF file for reading
	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("open input inputFile %s: %v", inputFilePath, err)
	}
	defer inputFile.Close()

	// Extract chapter information from PDF bookmarks
	chapters := extractChapters(inputFile)

	// Create separate PDF files for each chapter
	exportChapters(inputFile, chapters)
	return nil
}

// chapter represents a section in the PDF document.
// It contains the chapter title, order number, start page, and end page.
type chapter struct {
	title     string
	order     uint32
	startPage uint32
	endPage   uint32
}

// extractChapters reads the PDF bookmarks and converts them into chapter information.
// It filters out nested sub-chapters and keeps only top-level chapters.
// Parameters:
//   - inputFile: pointer to the opened PDF file
//
// Returns:
//   - []chapter: slice containing all chapter information
func extractChapters(inputFile *os.File) []chapter {
	// Create default configuration for PDF processing
	conf := model.NewDefaultConfiguration()

	// Extract bookmarks from the PDF file
	bookmarks, err := api.Bookmarks(inputFile, conf)
	if err != nil {
		log.Fatalf("failed to read PDF bookmarks: %v", err)
	}

	// Convert bookmarks to chapter information, skipping nested chapters
	var chapters []chapter
	for i, bm := range bookmarks {
		// Skip if this bookmark is within the page range of the previous chapter
		if len(chapters) > 0 && uint32(bm.PageFrom) < chapters[len(chapters)-1].endPage {
			continue
		}
		chapters = append(chapters, chapter{
			title:     bm.Title,
			order:     uint32(i + 1),
			startPage: uint32(bm.PageFrom),
		})
	}

	// Ensure at least one chapter was found
	if len(chapters) == 0 {
		log.Fatalf("no chapters found in input file")
	}

	// Set end pages for each chapter based on the next chapter's start page
	for i := 0; i < len(chapters)-1; i++ {
		chapters[i].endPage = chapters[i+1].startPage
	}

	// Set the end page of the last chapter to the total page count
	pageCount, err := api.PageCount(inputFile, conf)
	if err != nil {
		log.Fatalf("failed to read page count: %+v", err)
	}
	chapters[len(chapters)-1].endPage = uint32(pageCount)
	return chapters
}

// exportChapters creates separate PDF files for each chapter.
// Each chapter is saved as a separate PDF file with the format "order_chapterName.pdf".
// Parameters:
//   - inputFile: pointer to the source PDF file
//   - chapters: list of chapter information
func exportChapters(inputFile *os.File, chapters []chapter) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("fail to create output directory: %v", err)
	}

	// Process each chapter and create separate PDF files
	for _, cpt := range chapters {
		// Format the page range string for PDF splitting
		pageRange := fmt.Sprintf("%d-%d", cpt.startPage, cpt.endPage)

		// Generate output filename with chapter order and sanitized title
		outputFilePath := filepath.Join(outputDir, fmt.Sprintf("%02d_%s.pdf", cpt.order, sanitizeFilename(cpt.title)))

		// Create the output file
		outputFile, err := os.Create(outputFilePath)
		if err != nil {
			log.Fatalf("failed to create output file '%s': %v", outputFilePath, err)
		}

		// Extract the chapter pages to a new PDF file
		if err = api.Trim(inputFile, outputFile, []string{pageRange}, model.NewDefaultConfiguration()); err != nil {
			log.Fatalf("failed to split chapter '%s': %v", cpt.title, err)
		}
		fmt.Printf("exported chapter: '%s' (pages: %s)\n", cpt.title, pageRange)
	}
}

// sanitizeFilename cleans illegal characters from filename by replacing them with underscores.
// Common illegal characters include: /, \, :, *, ?, ", <, >, |
// Parameters:
//   - filename: original filename
//
// Returns:
//   - string: sanitized legal filename
func sanitizeFilename(filename string) string {
	// Define characters that are not allowed in filenames
	illegal := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := filename

	// Replace each illegal character with an underscore
	for _, char := range illegal {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}
