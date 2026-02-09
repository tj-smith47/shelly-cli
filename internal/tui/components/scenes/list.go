// Package scenes provides TUI components for managing scenes.
package scenes

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/errorview"
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

var footerHints = []keys.Hint{
	{Key: "a", Desc: "activate"},
	{Key: "v", Desc: "view"},
	{Key: "e", Desc: "edit"},
	{Key: "d", Desc: "del"},
	{Key: "n", Desc: "new"},
	{Key: "C", Desc: "capture"},
	{Key: "x", Desc: "export"},
	{Key: "i", Desc: "import"},
	{Key: "r", Desc: "refresh"},
}

// ListDeps holds the dependencies for the scenes list component.
type ListDeps struct {
	Ctx context.Context
	Svc *shelly.Service
}

// Validate ensures all required dependencies are set.
func (d ListDeps) Validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	if d.Svc == nil {
		return fmt.Errorf("service is required")
	}
	return nil
}

// LoadedMsg signals that scenes were loaded.
type LoadedMsg struct {
	Scenes []config.Scene
	Err    error
}

// ActionMsg signals a scene action result.
type ActionMsg struct {
	Action    string // "activate", "delete"
	SceneName string
	Err       error
}

// CreateSceneMsg signals that a new scene should be created.
type CreateSceneMsg struct{}

// CaptureSceneMsg signals that a scene should be captured from device states.
// The parent view handles this by collecting states from selected devices.
type CaptureSceneMsg struct{}

// EditSceneMsg signals that a scene should be edited.
type EditSceneMsg struct {
	Scene config.Scene
}

// ViewSceneMsg signals that a scene's details should be viewed.
type ViewSceneMsg struct {
	Scene config.Scene
}

// ExportSceneMsg signals that a scene was exported.
type ExportSceneMsg struct {
	SceneName string
	FilePath  string
	Err       error
}

// ImportSceneMsg signals that a scene was imported.
type ImportSceneMsg struct {
	SceneName string
	Err       error
}

// ListModel displays and manages scenes.
type ListModel struct {
	panel.Sizable
	ctx           context.Context
	svc           *shelly.Service
	scenes        []config.Scene
	loading       bool
	activating    bool
	err           error
	focused       bool
	panelIndex    int
	pendingDelete string // Scene name pending delete confirmation
	statusMsg     string // Temporary status message (export/import success)
	styles        ListStyles
}

// ListStyles holds styles for the list component.
type ListStyles struct {
	Name        lipgloss.Style
	Description lipgloss.Style
	ActionCount lipgloss.Style
	Selected    lipgloss.Style
	Error       lipgloss.Style
	Success     lipgloss.Style
	Muted       lipgloss.Style
	Cursor      lipgloss.Style
}

// DefaultListStyles returns the default styles for the scene list.
func DefaultListStyles() ListStyles {
	colors := theme.GetSemanticColors()
	return ListStyles{
		Name: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true),
		Description: lipgloss.NewStyle().
			Foreground(colors.Muted),
		ActionCount: lipgloss.NewStyle().
			Foreground(colors.Info),
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Success: lipgloss.NewStyle().
			Foreground(colors.Success),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Cursor: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
	}
}

// NewList creates a new scenes list model.
func NewList(deps ListDeps) ListModel {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("scenes list component init", err)
		panic(fmt.Sprintf("scenes: invalid deps: %v", err))
	}

	m := ListModel{
		Sizable: panel.NewSizable(4, panel.NewScroller(0, 10)),
		ctx:     deps.Ctx,
		svc:     deps.Svc,
		styles:  DefaultListStyles(),
	}
	m.Loader = m.Loader.SetMessage("Loading scenes...")
	return m
}

// Init returns the initial command.
func (m ListModel) Init() tea.Cmd {
	return m.loadScenes()
}

func (m ListModel) loadScenes() tea.Cmd {
	return func() tea.Msg {
		scenes := config.ListScenes()

		// Convert map to sorted slice
		result := make([]config.Scene, 0, len(scenes))
		for _, scene := range scenes {
			result = append(result, scene)
		}

		// Sort by name
		sort.Slice(result, func(i, j int) bool {
			return result[i].Name < result[j].Name
		})

		return LoadedMsg{Scenes: result}
	}
}

// SetSize sets the component dimensions.
func (m ListModel) SetSize(width, height int) ListModel {
	m.ApplySize(width, height)
	return m
}

// SetFocused sets the focus state.
func (m ListModel) SetFocused(focused bool) ListModel {
	m.focused = focused
	return m
}

// SetPanelIndex sets the 1-based panel index for Shift+N hotkey hint.
func (m ListModel) SetPanelIndex(index int) ListModel {
	m.panelIndex = index
	return m
}

// Update handles messages.
func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	// Forward tick messages to loader when loading
	if m.loading {
		result := generics.UpdateLoader(m.Loader, msg, func(msg tea.Msg) bool {
			switch msg.(type) {
			case LoadedMsg, ActionMsg:
				return true
			}
			return false
		})
		m.Loader = result.Loader
		if result.Consumed {
			return m, result.Cmd
		}
	}

	return m.handleMessage(msg)
}

func (m ListModel) handleMessage(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case LoadedMsg:
		return m.handleLoaded(msg)
	case ActionMsg:
		return m.handleAction(msg)
	case ExportSceneMsg:
		return m.handleExport(msg)
	case ImportSceneMsg:
		return m.handleImport(msg)
	case messages.NavigationMsg:
		return m.handleNavigationMsg(msg)
	case messages.ActivateRequestMsg, messages.ViewRequestMsg, messages.EditRequestMsg,
		messages.DeleteRequestMsg, messages.NewRequestMsg, messages.CaptureRequestMsg,
		messages.ExportRequestMsg, messages.ImportRequestMsg, messages.RefreshRequestMsg:
		return m.handleActionMsg(msg)
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m ListModel) handleNavigationMsg(msg messages.NavigationMsg) (ListModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	return m.handleNavigation(msg)
}

func (m ListModel) handleActionMsg(msg tea.Msg) (ListModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	m.pendingDelete = ""
	m.statusMsg = ""

	switch msg.(type) {
	case messages.ActivateRequestMsg:
		return m.activateScene()
	case messages.ViewRequestMsg:
		return m.viewScene()
	case messages.EditRequestMsg:
		return m.editScene()
	case messages.DeleteRequestMsg:
		return m.handleDelete()
	case messages.NewRequestMsg:
		return m, func() tea.Msg { return CreateSceneMsg{} }
	case messages.CaptureRequestMsg:
		return m, func() tea.Msg { return CaptureSceneMsg{} }
	case messages.ExportRequestMsg:
		return m.exportScene()
	case messages.ImportRequestMsg:
		return m.importScene()
	case messages.RefreshRequestMsg:
		m.loading = true
		return m, tea.Batch(m.Loader.Tick(), m.loadScenes())
	}
	return m, nil
}

func (m ListModel) handleLoaded(msg LoadedMsg) (ListModel, tea.Cmd) {
	m.loading = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.scenes = msg.Scenes
	m.Scroller.SetItemCount(len(m.scenes))
	return m, nil
}

func (m ListModel) handleAction(msg ActionMsg) (ListModel, tea.Cmd) {
	m.activating = false
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	// Refresh list after action
	return m, m.loadScenes()
}

func (m ListModel) handleExport(msg ExportSceneMsg) (ListModel, tea.Cmd) {
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.statusMsg = fmt.Sprintf("Exported %q to %s", msg.SceneName, msg.FilePath)
	return m, nil
}

func (m ListModel) handleImport(msg ImportSceneMsg) (ListModel, tea.Cmd) {
	if msg.Err != nil {
		m.err = msg.Err
		return m, nil
	}
	m.statusMsg = fmt.Sprintf("Imported scene %q", msg.SceneName)
	// Refresh list after import
	return m, m.loadScenes()
}

func (m ListModel) handleNavigation(msg messages.NavigationMsg) (ListModel, tea.Cmd) {
	m.pendingDelete = "" // Clear pending delete on navigation
	m.Scroller.HandleNavigation(msg)
	return m, nil
}

func (m ListModel) handleDelete() (ListModel, tea.Cmd) {
	scene := m.selectedScene()
	if scene == nil {
		return m, nil
	}
	if m.pendingDelete == scene.Name {
		// Second press - confirm delete
		m.pendingDelete = ""
		return m.deleteScene()
	}
	// First press - mark pending
	m.pendingDelete = scene.Name
	return m, nil
}

func (m ListModel) handleKey(msg tea.KeyPressMsg) (ListModel, tea.Cmd) {
	// Handle component-specific keys not covered by action messages
	if msg.String() == "esc" && m.pendingDelete != "" {
		// Cancel pending delete
		m.pendingDelete = ""
		return m, nil
	}
	return m, nil
}

func (m ListModel) selectedScene() *config.Scene {
	cursor := m.Scroller.Cursor()
	if len(m.scenes) == 0 || cursor >= len(m.scenes) {
		return nil
	}
	return &m.scenes[cursor]
}

func (m ListModel) activateScene() (ListModel, tea.Cmd) {
	scene := m.selectedScene()
	if scene == nil {
		return m, nil
	}

	m.activating = true
	sceneCopy := *scene

	return m, func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 60*time.Second)
		defer cancel()

		// Execute each action in the scene
		var errors []string
		for _, action := range sceneCopy.Actions {
			if err := executeAction(ctx, m.svc, action); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", action.Device, err))
			}
		}

		if len(errors) > 0 {
			return ActionMsg{
				Action:    "activate",
				SceneName: sceneCopy.Name,
				Err:       fmt.Errorf("some actions failed: %s", strings.Join(errors, "; ")),
			}
		}

		return ActionMsg{Action: "activate", SceneName: sceneCopy.Name}
	}
}

// executeAction executes a single scene action on a device.
func executeAction(ctx context.Context, svc *shelly.Service, action config.SceneAction) error {
	_, err := svc.RawRPC(ctx, action.Device, action.Method, action.Params)
	return err
}

func (m ListModel) editScene() (ListModel, tea.Cmd) {
	scene := m.selectedScene()
	if scene == nil {
		return m, nil
	}
	sceneCopy := *scene
	return m, func() tea.Msg {
		return EditSceneMsg{Scene: sceneCopy}
	}
}

func (m ListModel) viewScene() (ListModel, tea.Cmd) {
	scene := m.selectedScene()
	if scene == nil {
		return m, nil
	}
	sceneCopy := *scene
	return m, func() tea.Msg {
		return ViewSceneMsg{Scene: sceneCopy}
	}
}

func (m ListModel) deleteScene() (ListModel, tea.Cmd) {
	scene := m.selectedScene()
	if scene == nil {
		return m, nil
	}
	sceneName := scene.Name

	return m, func() tea.Msg {
		err := config.DeleteScene(sceneName)
		return ActionMsg{Action: "delete", SceneName: sceneName, Err: err}
	}
}

func (m ListModel) exportScene() (ListModel, tea.Cmd) {
	scene := m.selectedScene()
	if scene == nil {
		return m, nil
	}
	sceneName := scene.Name

	return m, func() tea.Msg {
		// Get scenes export directory
		configDir, err := config.Dir()
		if err != nil {
			return ExportSceneMsg{SceneName: sceneName, Err: err}
		}

		scenesDir := filepath.Join(configDir, "scenes")
		if err := config.Fs().MkdirAll(scenesDir, 0o755); err != nil {
			return ExportSceneMsg{SceneName: sceneName, Err: err}
		}

		// Export to file
		outputPath := filepath.Join(scenesDir, sceneName+".json")
		filePath, err := config.ExportSceneToFile(sceneName, outputPath)
		return ExportSceneMsg{SceneName: sceneName, FilePath: filePath, Err: err}
	}
}

func (m ListModel) importScene() (ListModel, tea.Cmd) {
	return m, func() tea.Msg {
		// Get scenes import directory
		configDir, err := config.Dir()
		if err != nil {
			return ImportSceneMsg{Err: err}
		}

		scenesDir := filepath.Join(configDir, "scenes")

		// List available scene files
		entries, err := afero.ReadDir(config.Fs(), scenesDir)
		if err != nil {
			return ImportSceneMsg{Err: fmt.Errorf("no scenes directory: %s", scenesDir)}
		}

		// Find first JSON/YAML file that's not already imported
		existingScenes := config.ListScenes()
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			ext := filepath.Ext(name)
			if ext != ".json" && ext != ".yaml" && ext != ".yml" {
				continue
			}
			// Check if scene already exists
			baseName := strings.TrimSuffix(name, ext)
			if _, exists := existingScenes[baseName]; exists {
				continue
			}
			// Import this scene
			filePath := filepath.Join(scenesDir, name)
			msg, err := config.ImportSceneFromFile(filePath, "", false)
			if err != nil {
				return ImportSceneMsg{Err: err}
			}
			// Parse scene name from message
			return ImportSceneMsg{SceneName: msg}
		}

		return ImportSceneMsg{Err: fmt.Errorf("no new scene files found in %s", scenesDir)}
	}
}

// View renders the scenes list.
func (m ListModel) View() string {
	r := rendering.New(m.Width, m.Height).
		SetTitle("Scenes").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Add footer with keybindings when focused
	if m.focused && len(m.scenes) > 0 {
		footer := m.buildFooter()
		r.SetFooter(footer)
	}

	if m.loading {
		r.SetContent(m.Loader.View())
		return r.Render()
	}

	if m.activating {
		r.SetContent(m.styles.Muted.Render("Activating scene..."))
		return r.Render()
	}

	if m.err != nil {
		if tuierrors.IsUnsupportedFeature(m.err) {
			r.SetContent(styles.EmptyStateWithBorder(tuierrors.UnsupportedMessage("Scenes"), m.Width, m.Height))
		} else {
			r.SetContent(errorview.RenderInline(m.err))
		}
		return r.Render()
	}

	if len(m.scenes) == 0 {
		r.SetContent(styles.EmptyStateWithBorder("No scenes defined\nPress n to create one", m.Width, m.Height))
		return r.Render()
	}

	content := generics.RenderScrollableList(generics.ListRenderConfig[config.Scene]{
		Items:    m.scenes,
		Scroller: m.Scroller,
		RenderItem: func(scene config.Scene, _ int, isCursor bool) string {
			return m.renderSceneLine(scene, isCursor)
		},
		ScrollStyle:    m.styles.Muted,
		ScrollInfoMode: generics.ScrollInfoAlways,
	})

	r.SetContent(content)
	return r.Render()
}

func (m ListModel) renderSceneLine(scene config.Scene, isSelected bool) string {
	// Selection indicator
	selector := "  "
	if isSelected && m.focused {
		selector = m.styles.Cursor.Render("> ")
	}

	// Action count
	actionCount := fmt.Sprintf("(%d actions)", len(scene.Actions))

	// Calculate available width for name
	// Fixed: selector(2) + space(1) + actionCount length
	available := output.ContentWidth(m.Width, 4+3+len(actionCount))
	name := output.Truncate(scene.Name, max(available, 10))

	line := fmt.Sprintf("%s%s %s",
		selector,
		m.styles.Name.Render(name),
		m.styles.ActionCount.Render(actionCount),
	)

	if isSelected && m.focused {
		return m.styles.Selected.Render(line)
	}
	return line
}

func (m ListModel) buildFooter() string {
	// Show delete confirmation message if pending
	if m.pendingDelete != "" {
		return m.styles.Error.Render("Press d again to delete, Esc to cancel")
	}

	// Show status message if set
	if m.statusMsg != "" {
		return m.styles.Success.Render(m.statusMsg)
	}

	return theme.StyledKeybindings(keys.FormatHints(footerHints, keys.FooterHintWidth(m.Width)))
}

// SelectedScene returns the currently selected scene, if any.
func (m ListModel) SelectedScene() *config.Scene {
	return m.selectedScene()
}

// Cursor returns the current cursor position.
func (m ListModel) Cursor() int {
	return m.Scroller.Cursor()
}

// SceneCount returns the number of scenes.
func (m ListModel) SceneCount() int {
	return len(m.scenes)
}

// Loading returns whether the component is loading.
func (m ListModel) Loading() bool {
	return m.loading
}

// Activating returns whether a scene is being activated.
func (m ListModel) Activating() bool {
	return m.activating
}

// Error returns any error that occurred.
func (m ListModel) Error() error {
	return m.err
}

// Refresh triggers a refresh of the scene list.
func (m ListModel) Refresh() (ListModel, tea.Cmd) {
	m.loading = true
	return m, tea.Batch(m.Loader.Tick(), m.loadScenes())
}

// FooterText returns keybinding hints for the footer.
func (m ListModel) FooterText() string {
	return keys.FormatHints(footerHints, 0)
}
