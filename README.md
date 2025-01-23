# PDF Chapter Splitter

A command-line tool that splits PDF files into separate chapters based on their table of contents (bookmarks). Ideal for breaking down large PDF books or documents into smaller, chapter-based files.

## Features

- 📚 Automatically splits PDF by bookmarks
- 🎯 Focuses on top-level chapters
- 🔄 Maintains original PDF quality
- 📝 Auto-sanitizes chapter names for filenames
- 📂 Creates organized, numbered chapter files

## Installation

```bash
go install github.com/souhup/pdf-spliter
```

## Command Line Options

| Flag | Description | Required | Default |
|------|-------------|----------|---------|
| `-i, --input` | Input PDF file path | Yes | - |
| `-o, --output` | Output directory | No | "output" |

## Technical Details

The tool works by:
1. Reading the PDF's bookmark structure
2. Identifying top-level chapters
3. Creating separate PDF files for each chapter
4. Naming files with chapter numbers and sanitized titles

## Limitations

- Requires PDF files with table of contents (bookmarks)
- Only processes top-level bookmarks
- Skips nested sub-chapters
- Chapter titles must be unique after sanitization

## Dependencies

- [pdfcpu](https://github.com/pdfcpu/pdfcpu) 
- [cobra](https://github.com/spf13/cobra) 

## Support

If you encounter any issues or have questions, please [open an issue](https://github.com/souhup/pdf-spliter/issues).