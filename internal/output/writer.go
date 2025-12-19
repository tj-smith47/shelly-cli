package output

import (
	"fmt"
	"io"
	"os"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// GetWriter returns a writer for the specified output file.
// If outputFile is empty, returns stdout. Otherwise creates the file.
// Returns the writer, a closer function, and any error.
func GetWriter(ios *iostreams.IOStreams, outputFile string) (io.Writer, func(), error) {
	if outputFile == "" {
		return ios.Out, func() {}, nil
	}

	//nolint:gosec // G304: User-provided file path is expected for CLI export functionality
	file, err := os.Create(outputFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create output file: %w", err)
	}

	return file, func() {
		if err := file.Close(); err != nil {
			ios.DebugErr("close output file", err)
		}
	}, nil
}

// ExportToFile exports data to a file using the specified format.
// If outputFile is empty, outputs to stdout without success message.
func ExportToFile(ios *iostreams.IOStreams, data any, outputFile string, format Format, formatName string) error {
	writer, closer, err := GetWriter(ios, outputFile)
	if err != nil {
		return err
	}
	defer closer()

	formatter := NewFormatter(format)
	if err := formatter.Format(writer, data); err != nil {
		return fmt.Errorf("failed to encode %s: %w", formatName, err)
	}

	if outputFile != "" {
		ios.Success("Exported to %s (%s)", outputFile, formatName)
	}
	return nil
}

// ExportCSV exports CSV data to a file using a formatter function.
// If outputFile is empty, outputs to stdout without success message.
func ExportCSV(ios *iostreams.IOStreams, outputFile string, formatter func() ([]byte, error)) error {
	csvData, err := formatter()
	if err != nil {
		return err
	}

	writer, closer, err := GetWriter(ios, outputFile)
	if err != nil {
		return err
	}
	defer closer()

	if _, err := writer.Write(csvData); err != nil {
		return fmt.Errorf("failed to write CSV data: %w", err)
	}

	if outputFile != "" {
		ios.Success("Exported to %s (CSV)", outputFile)
	}
	return nil
}
