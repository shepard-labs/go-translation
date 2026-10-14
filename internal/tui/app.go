package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepard-labs/go-translation/internal/config"
	"github.com/shepard-labs/go-translation/internal/providers"
	"github.com/shepard-labs/go-translation/internal/translate"
)

type model struct {
	ctx       context.Context
	cfg       config.Config
	provider  textinput.Model
	modelName textinput.Model
	from      textinput.Model
	to        textinput.Model
	input     textarea.Model
	spinner   spinner.Model
	focus     int
	busy      bool
	output    string
	err       error
	status    string
}

type translatedMsg struct {
	text string
	err  error
}

func Run(ctx context.Context, cfg config.Config) error {
	_, err := tea.NewProgram(newModel(ctx, cfg), tea.WithAltScreen()).Run()
	return err
}

func newModel(ctx context.Context, cfg config.Config) model {
	provider := textinput.New()
	provider.Placeholder = "anthropic"
	provider.SetValue(cfg.DefaultProvider)
	provider.Focus()

	modelName := textinput.New()
	modelName.Placeholder = "model"
	modelName.SetValue(config.ModelFor(cfg))

	from := textinput.New()
	from.Placeholder = "auto"
	from.SetValue(cfg.DefaultSourceLanguage)

	to := textinput.New()
	to.Placeholder = "Spanish"
	to.SetValue(cfg.DefaultTargetLanguage)

	inputArea := textarea.New()
	inputArea.Placeholder = "Paste text to translate"
	inputArea.SetWidth(80)
	inputArea.SetHeight(8)

	spin := spinner.New()
	spin.Spinner = spinner.Dot

	return model{ctx: ctx, cfg: cfg, provider: provider, modelName: modelName, from: from, to: to, input: inputArea, spinner: spin}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.nextFocus()
			return m, nil
		case "shift+tab":
			m.prevFocus()
			return m, nil
		case "ctrl+t":
			m.busy = true
			m.err = nil
			m.status = "translating"
			return m, tea.Batch(m.spinner.Tick, m.translateCmd())
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if m.busy {
			return m, cmd
		}
		return m, nil
	case translatedMsg:
		m.busy = false
		m.err = msg.err
		m.output = msg.text
		m.status = "done"
		if msg.err != nil {
			m.status = "error"
		}
		return m, nil
	}

	var cmd tea.Cmd
	switch m.focus {
	case 0:
		m.provider, cmd = m.provider.Update(msg)
	case 1:
		m.modelName, cmd = m.modelName.Update(msg)
	case 2:
		m.from, cmd = m.from.Update(msg)
	case 3:
		m.to, cmd = m.to.Update(msg)
	case 4:
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	title := lipgloss.NewStyle().Bold(true).Render("aitranslate TUI")
	status := m.status
	if m.busy {
		status = m.spinner.View() + " translating"
	}
	if m.err != nil {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(m.err.Error())
	}
	return strings.Join([]string{
		title,
		"tab/shift+tab move | ctrl+t translate | q quit",
		"Provider: " + m.provider.View(),
		"Model: " + m.modelName.View(),
		"From: " + m.from.View(),
		"To: " + m.to.View(),
		"Input:",
		m.input.View(),
		"Status: " + status,
		"Output:",
		m.output,
	}, "\n")
}

func (m *model) nextFocus() {
	m.setFocus((m.focus + 1) % 5)
}

func (m *model) prevFocus() {
	m.setFocus((m.focus + 4) % 5)
}

func (m *model) setFocus(next int) {
	m.provider.Blur()
	m.modelName.Blur()
	m.from.Blur()
	m.to.Blur()
	m.input.Blur()
	m.focus = next
	switch next {
	case 0:
		m.provider.Focus()
	case 1:
		m.modelName.Focus()
	case 2:
		m.from.Focus()
	case 3:
		m.to.Focus()
	case 4:
		m.input.Focus()
	}
}

func (m model) translateCmd() tea.Cmd {
	return func() tea.Msg {
		cfg := m.cfg
		cfg.DefaultProvider = m.provider.Value()
		cfg.DefaultModel = m.modelName.Value()
		cfg.DefaultSourceLanguage = m.from.Value()
		cfg.DefaultTargetLanguage = m.to.Value()
		if err := config.Validate(cfg); err != nil {
			return translatedMsg{err: err}
		}
		translator, err := providers.New(cfg)
		if err != nil {
			return translatedMsg{err: err}
		}
		service := translate.NewService(translator, cfg.Chunking.Enabled, cfg.Chunking.MaxChars)
		result, err := service.Translate(m.ctx, translate.TranslationRequest{
			Provider:        cfg.DefaultProvider,
			Model:           config.ModelFor(cfg),
			SourceLanguage:  cfg.DefaultSourceLanguage,
			TargetLanguage:  cfg.DefaultTargetLanguage,
			Text:            m.input.Value(),
			MaxOutputTokens: cfg.MaxOutputTokens,
			Temperature:     cfg.Temperature,
		})
		if err != nil {
			return translatedMsg{err: err}
		}
		return translatedMsg{text: result.Text}
	}
}

func (m model) String() string {
	return fmt.Sprintf("provider=%s model=%s", m.provider.Value(), m.modelName.Value())
}
