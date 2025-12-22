# Theme System Guide

The Shelly CLI supports 280+ built-in themes via [bubbletint](https://github.com/lrstanley/bubbletint), with the ability to customize colors and create your own themes.

## Quick Start

```bash
# List available themes
shelly theme list

# Set a theme
shelly theme set dracula

# Preview a theme
shelly theme preview nord

# Show current theme
shelly theme current
```

## Built-in Themes

The CLI includes 280+ themes from bubbletint. Here are some popular options:

### Dark Themes

| Theme | Description |
|-------|-------------|
| `dracula` | Dark theme with vibrant purple, pink, and green (default) |
| `nord` | Arctic, north-bluish color palette |
| `gruvbox-dark` | Warm, retro groove colors |
| `tokyo-night` | Modern dark theme inspired by Tokyo |
| `catppuccin-mocha` | Soothing pastel theme - Mocha variant |
| `one-dark` | Atom One Dark color scheme |
| `monokai` | Classic syntax highlighting theme |
| `material` | Material Design inspired colors |
| `solarized-dark` | Precision colors for machines and people |
| `ayu-dark` | Simple, bright colors with dark background |

### Light Themes

| Theme | Description |
|-------|-------------|
| `solarized-light` | Light variant of Solarized |
| `catppuccin-latte` | Light pastel theme - Latte variant |
| `gruvbox-light` | Light variant of Gruvbox |
| `one-light` | Atom One Light color scheme |
| `ayu-light` | Light variant of Ayu |

### Specialized Themes

| Theme | Description |
|-------|-------------|
| `cyberpunk` | Neon-inspired cyberpunk aesthetic |
| `synthwave` | 80s retrowave inspired |
| `github-dark` | GitHub's dark theme |
| `vs-code` | VS Code default dark theme |
| `sublime` | Sublime Text inspired |

## Setting a Theme

### Via Command

```bash
shelly theme set <theme-name>
```

### Via Configuration

```yaml
# ~/.config/shelly/config.yaml
theme: dracula
```

### Via Environment Variable

```bash
export SHELLY_THEME=nord
```

## Theme Configuration

The theme setting supports three formats:

### 1. Simple Theme Name

```yaml
theme: dracula
```

### 2. Theme with Color Overrides

Override specific colors while keeping the base theme:

```yaml
theme:
  name: dracula
  colors:
    green: "#00ff00"    # Override success color
    red: "#ff0000"      # Override error color
```

### 3. External Theme File

Load theme from a file:

```yaml
theme:
  file: ~/.config/shelly/themes/mytheme.yaml
```

## Custom Themes

### Creating a Custom Theme

Create a YAML file in `~/.config/shelly/themes/`:

```yaml
# ~/.config/shelly/themes/mytheme.yaml
name: mytheme
colors:
  foreground: "#f8f8f2"
  background: "#282a36"
  green: "#50fa7b"
  red: "#ff5555"
  yellow: "#f1fa8c"
  blue: "#6272a4"
  cyan: "#8be9fd"
  purple: "#bd93f9"
  bright_black: "#44475a"
```

### Available Color Properties

| Property | Description | Usage |
|----------|-------------|-------|
| `foreground` | Default text color | Labels, general text |
| `background` | Background color | Table backgrounds |
| `green` | Success color | OK status, enabled states |
| `red` | Error color | Errors, disabled states |
| `yellow` | Warning color | Warnings, updating states |
| `blue` | Info color | Information, links |
| `cyan` | Highlight color | Values, emphasis |
| `purple` | Accent color | Titles, headings |
| `bright_black` | Dim/muted color | Subtitles, secondary text |

### Using Your Theme

```yaml
# ~/.config/shelly/config.yaml
theme:
  file: ~/.config/shelly/themes/mytheme.yaml
```

Or reference by name if in themes directory:

```bash
shelly theme set mytheme
```

## TUI-Specific Themes

The TUI dashboard can use a different theme than the CLI:

```yaml
# ~/.config/shelly/config.yaml
theme: dracula  # CLI theme

tui:
  theme:
    name: nord  # Independent TUI theme
```

This allows using a compact theme for CLI output while having a different aesthetic for the dashboard.

## Color Overrides

Override individual colors without changing the base theme:

```yaml
theme:
  name: dracula
  colors:
    green: "#00ff88"   # Brighter green for success
    cyan: "#00ccff"    # Different highlight color
```

## Theme Preview

Preview how a theme looks before setting it:

```bash
# Preview a built-in theme
shelly theme preview nord

# The preview shows:
# - Color palette
# - Sample output with the theme applied
# - How different elements appear
```

## Listing Themes

```bash
# List all available themes
shelly theme list

# Filter themes by name
shelly theme list --filter dark

# Show themes with descriptions
shelly theme list --verbose
```

## Theme Categories

Themes are organized into categories:

| Category | Description |
|----------|-------------|
| Dark | Dark background themes |
| Light | Light background themes |
| Pastel | Soft, muted color themes |
| Vibrant | High contrast, colorful themes |
| Retro | Vintage-inspired themes |
| Material | Material Design based |
| Syntax | Based on code editor themes |

## How Themes Are Applied

Themes affect all CLI output:

### Status Indicators

```
# With dracula theme:
  online   (green)
  offline  (red)
  updating (yellow)
```

### Device Status

```
Device: kitchen-light
Status: ON         (green)
Power:  45.2W      (cyan)
Energy: 1.2kWh     (blue)
```

### Tables

```
NAME          ADDRESS          STATUS
living-room   192.168.1.100   online   (green)
kitchen       192.168.1.101   online   (green)
garage        192.168.1.102   offline  (red)
```

### Progress Indicators

```
 Processing devices... (spinner in theme accent color)
```

## Disabling Colors

For scripts or piping, disable colors:

```bash
# Via flag
shelly device list --no-color

# Via environment
export SHELLY_NO_COLOR=1
shelly device list

# NO_COLOR standard
export NO_COLOR=1
shelly device list
```

## Semantic Colors

The theme system includes semantic color roles that provide consistent meaning across the entire CLI. These colors are automatically mapped from the selected theme but can be overridden in your config.

### Available Semantic Colors

| Role | Purpose | Default Mapping |
|------|---------|-----------------|
| `primary` | Main actions, highlights | Purple |
| `secondary` | Supporting UI elements | Blue |
| `highlight` | Emphasis, selection | Cyan |
| `muted` | Disabled, less important | Gray (bright_black) |
| `text` | Primary text | Foreground |
| `alt_text` | Secondary text | Gray (bright_black) |
| `success` | Successful operations | Green |
| `warning` | Warnings, cautions | Yellow |
| `error` | Errors, failures | Red |
| `info` | Informational messages | Cyan |
| `background` | Main background | Background |
| `alt_background` | Alternate background | Gray (bright_black) |
| `online` | Device online status | Green |
| `offline` | Device offline status | Red |
| `updating` | Device updating | Yellow |
| `idle` | Device idle/inactive | Gray |
| `table_header` | Table headers | Cyan |
| `table_cell` | Table cells | Yellow |
| `table_alt_cell` | Alternating table cells | Orange |
| `table_border` | Table borders | Purple |

### Viewing Current Semantic Colors

Use the `theme semantic` command to see how colors are mapped:

```bash
shelly theme semantic
```

This displays a color swatch for each semantic role.

### Overriding Semantic Colors

Add a `semantic` section to your config to customize specific colors:

```yaml
# ~/.config/shelly/config.yaml
theme:
  name: dracula
  semantic:
    primary: "#ff79c6"      # Use pink instead of purple
    success: "#8be9fd"      # Use cyan instead of green
    table_header: "#ffb86c" # Orange table headers
```

Only the colors you specify are overridden; the rest use the theme defaults.

### Theme-Specific Mappings

Popular themes have curated semantic mappings that match their aesthetic:

- **Dracula**: Purple primary, pink accents, classic Dracula palette
- **Nord**: Frost cyan primary, aurora colors for feedback
- **Tokyo Night**: Blue primary, soft pastels throughout
- **Gruvbox**: Gold primary, warm earthy tones
- **Catppuccin**: Mauve primary, soothing pastel colors

Other themes use a generic mapping that reads colors directly from the theme palette.

## Theme API (For Developers)

The theme package provides these functions:

```go
import "github.com/tj-smith47/shelly-cli/internal/theme"

// Get raw colors (color.Color)
theme.Fg()      // Foreground color
theme.Bg()      // Background color
theme.Green()   // Green color
theme.Red()     // Red color
theme.Yellow()  // Yellow color
theme.Blue()    // Blue color
theme.Cyan()    // Cyan color
theme.Purple()  // Purple color

// Get semantic colors (preferred for consistent theming)
colors := theme.GetSemanticColors()
colors.Primary      // Main action color
colors.Success      // Success/ok operations
colors.Error        // Error/failure operations
colors.Warning      // Warning messages
colors.Online       // Device online state
colors.Offline      // Device offline state
colors.TableHeader  // Table header color
colors.TableCell    // Table cell color

// Semantic styles (lipgloss.Style) - use these for consistency
theme.SemanticPrimary()    // Primary actions
theme.SemanticSuccess()    // Success feedback
theme.SemanticError()      // Error feedback
theme.SemanticWarning()    // Warning feedback
theme.SemanticInfo()       // Informational
theme.SemanticOnline()     // Device online
theme.SemanticOffline()    // Device offline

// Legacy status styles (now use semantic colors internally)
theme.StatusOK()      // For success status (uses Success semantic)
theme.StatusWarn()    // For warnings (uses Warning semantic)
theme.StatusError()   // For errors (uses Error semantic)
theme.StatusInfo()    // For information (uses Info semantic)
theme.StatusOnline()  // For device online (uses Online semantic)
theme.StatusOffline() // For device offline (uses Offline semantic)

// Other styles
theme.Bold()          // Bold text
theme.Dim()           // Dimmed text
theme.Title()         // For titles
theme.Link()          // For URLs

// Pre-rendered strings
theme.DeviceOnline()   // "  online"
theme.DeviceOffline()  // "  offline"
theme.SwitchOn()       // "ON"
theme.SwitchOff()      // "OFF"
```

## Troubleshooting

### Theme Not Applying

1. Check theme name spelling: `shelly theme list | grep <name>`
2. Verify configuration syntax
3. Restart terminal to clear cached styles

### Colors Look Wrong

1. Ensure terminal supports 256 colors or true color
2. Check terminal color settings
3. Try `export COLORTERM=truecolor`

### Custom Theme Not Loading

1. Verify file path is correct
2. Check YAML syntax: `cat ~/.config/shelly/themes/mytheme.yaml | python -c "import yaml, sys; yaml.safe_load(sys.stdin)"`
3. Ensure all required color properties are set

### No Colors in Pipe

Colors are auto-disabled when stdout is not a TTY. Use `--force-color` to override if needed.

## Examples

### Corporate Theme

```yaml
# ~/.config/shelly/themes/corporate.yaml
name: corporate
colors:
  foreground: "#333333"
  background: "#ffffff"
  green: "#28a745"
  red: "#dc3545"
  yellow: "#ffc107"
  blue: "#007bff"
  cyan: "#17a2b8"
  purple: "#6f42c1"
  bright_black: "#6c757d"
```

### High Contrast Theme

```yaml
# ~/.config/shelly/themes/high-contrast.yaml
name: high-contrast
colors:
  foreground: "#ffffff"
  background: "#000000"
  green: "#00ff00"
  red: "#ff0000"
  yellow: "#ffff00"
  blue: "#0000ff"
  cyan: "#00ffff"
  purple: "#ff00ff"
  bright_black: "#808080"
```

### Matching Terminal Theme

If your terminal uses a specific theme (e.g., Dracula), set the same in the CLI for a consistent look:

```yaml
theme: dracula
```

## Resources

- [bubbletint themes](https://github.com/lrstanley/bubbletint) - Source of 280+ themes
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Styling library
- [Terminal.sexy](https://terminal.sexy/) - Create color schemes
- [Gogh](https://mayccoll.github.io/Gogh/) - Terminal color schemes
