package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"github.com/moolite/bot/internal/llm"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Padding(0, 1)

	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("213")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("86")).
			Padding(0, 2)
)

type model struct {
	messages []llm.Message
	input    textinput.Model
	loading  bool
	useTools bool
	err      error
	client   *llm.Generator
}

type llmResponseMsg struct {
	resp *llm.Message
	err  error
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type a message..."
	ti.Focus()
	ti.CharLimit = 2000
	ti.SetWidth(60)
	ti.ShowSuggestions = true
	ti.SetSuggestions([]string{"/exit", "/quit", "/reset", "/context"})

	client, err := llm.NewClient(context.Background(), -999)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create LLM client: %v\n", err)
		os.Exit(1)
	}

	return model{
		messages: []llm.Message{{Role: "system", Content: llm.SystemPrompt}},
		input:    ti,
		client:   client,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q", "ctrl+d", "esc":
			return m, tea.Quit
		case "ctrl+y":
			m.useTools = !m.useTools
			return m, nil
		case "enter":
			return m.handleSend()
		}

	case llmResponseMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.messages = append(m.messages, *msg.resp)
		m.err = nil
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *model) handleSend() (tea.Model, tea.Cmd) {
	text := strings.TrimSpace(m.input.Value())
	if text == "" {
		return m, nil
	}

	if text == "/exit" || text == "/quit" {
		return m, tea.Quit
	}

	if text == "/reset" {
		m.messages = []llm.Message{{Role: "system", Content: llm.SystemPrompt}}
		m.input.SetValue("")
		m.err = nil
		return m, nil
	}

	if text == "/context" {
		m.input.SetValue("")
		return m, nil
	}

	m.messages = append(m.messages, llm.Message{Role: "user", Content: text})
	m.input.SetValue("")
	m.loading = true
	m.err = nil

	return m, func() tea.Msg {
		ctx := context.Background()
		var resp *llm.Message
		var err error
		if m.useTools {
			resp, err = m.client.Chat(ctx, m.messages, llm.AllTools())
		} else {
			resp, err = m.client.Chat(ctx, m.messages, []llm.Tool{})
		}
		return llmResponseMsg{resp: resp, err: err}
	}
}

func (m model) View() tea.View {
	var b strings.Builder

	b.WriteString(headerStyle.Render("LLM Chat"))
	b.WriteString("  ")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("[/reset] [/exit]"))
	if m.useTools {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Padding(0, 1).Bold(true).Render("(ctrl+y tools)"))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 1).Render("[ctrl+y tools]"))
	}
	b.WriteString("\n\n")

	for i, msg := range m.messages {
		if msg.Role == "system" {
			continue
		}

		var style lipgloss.Style
		var prefix string
		switch msg.Role {
		case "user":
			style = userStyle
			prefix = "You"
		case "assistant":
			style = assistantStyle
			prefix = "Marrano"
		case "tool":
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			prefix = "Tool"
		default:
			style = lipgloss.NewStyle()
			prefix = msg.Role
		}

		b.WriteString(style.Render(fmt.Sprintf("[%s]", prefix)))
		b.WriteString("  ")
		b.WriteString(msg.Content)
		if i < len(m.messages)-1 {
			b.WriteString("\n\n")
		}
	}

	if m.loading {
		b.WriteString("\n\n")
		b.WriteString(loadingStyle.Render("Thinking..."))
	}

	if m.err != nil {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	b.WriteString("\n\n")
	b.WriteString(inputStyle.Render(m.input.View()))

	return tea.NewView(b.String())
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
