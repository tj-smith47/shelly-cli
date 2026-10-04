package iostreams

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
)

// Styles always render escape codes, so the process streams must strip them
// whenever colour is off: --no-color, NO_COLOR, --plain or a pipe.
func TestProcessStreamsStripEscapesWhenColorIsOff(t *testing.T) {
	t.Parallel()

	const styled = "\x1b[38;2;255;85;85mDisabled\x1b[m"
	var out, errOut bytes.Buffer
	ios := &IOStreams{
		outColor: &colorprofile.Writer{Forward: &out, Profile: colorprofile.TrueColor},
		errColor: &colorprofile.Writer{Forward: &errOut, Profile: colorprofile.TrueColor},
	}
	ios.Out, ios.ErrOut = ios.outColor, ios.errColor

	ios.SetColorEnabled(true)
	ios.Printf("%s", styled)
	if !strings.Contains(out.String(), "\x1b[") {
		t.Fatalf("colour on: stdout = %q, want the escape codes kept", out.String())
	}

	out.Reset()
	ios.SetColorEnabled(false)
	ios.Printf("%s", styled)
	ios.Errorf("%s", styled)
	if out.String() != "Disabled" || errOut.String() != "Disabled" {
		t.Errorf("colour off: stdout = %q, stderr = %q, want plain text on both", out.String(), errOut.String())
	}
}
