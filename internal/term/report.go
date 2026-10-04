package term

import (
	"bytes"
	"fmt"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/jsonfmt"
	"github.com/tj-smith47/shelly-cli/internal/output/yamlfmt"
)

// OutputReport writes a report as json, yaml or text (table is the global
// -o spelling of text) to stdout, or to outputPath when it is set. text
// renders the text format. A file never carries terminal colors.
func OutputReport[T any](ios *iostreams.IOStreams, report T, text func(T) string, format, outputPath string) error {
	var buf bytes.Buffer
	switch format {
	case string(output.FormatJSON):
		f := jsonfmt.New()
		f.Highlight = f.Highlight && outputPath == ""
		if err := f.Format(&buf, report); err != nil {
			return fmt.Errorf("failed to format report: %w", err)
		}
	case string(output.FormatYAML):
		f := yamlfmt.New()
		f.Highlight = f.Highlight && outputPath == ""
		if err := f.Format(&buf, report); err != nil {
			return fmt.Errorf("failed to format report: %w", err)
		}
	case string(output.FormatText), string(output.FormatTable):
		buf.WriteString(text(report))
	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	if outputPath != "" {
		if err := afero.WriteFile(config.Fs(), outputPath, buf.Bytes(), 0o600); err != nil {
			return fmt.Errorf("failed to write report: %w", err)
		}
		ios.Success("Report saved to: %s", outputPath)
		return nil
	}

	ios.Printf("%s", buf.String())
	return nil
}
