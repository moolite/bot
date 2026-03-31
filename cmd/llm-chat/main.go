package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/moolite/bot/internal/db"
	"github.com/moolite/bot/internal/embedder"
	"github.com/moolite/bot/internal/llm"
	"github.com/moolite/bot/internal/vectorstore"
)

const maxMessages = 100

var (
	petnameAdjectives = []string{
		"grumpy", "happy", "sleepy", "bouncy", "fluffy", "sneaky", "brave", "shy",
		"quick", "jolly", "bold", "calm", "crazy", "dizzy", "eager", "fancy",
		"gentle", "hasty", "idle", "jumpy", "kind", "lively", "merry", "nervous",
		"odd", "proud", "quiet", "rough", "silly", "tidy", "uptight", "vivid",
		"witty", "zany", "angry", "bright", "clever", "daring", "easily", "foolish",
		"generous", "honest", "inner", "joyful", "keen", "lovely", "mighty", "noble",
		"open", "polite", "quaint", "real", "stupid", "trusty", "useful", "valiant",
		"whole", "youthful", "zealous",
	}
	petnameAnimals = []string{
		"frog", "panda", "cat", "dog", "owl", "bear", "wolf", "fox",
		"hawk", "lion", "tiger", "eagle", "rabbit", "deer", "snake", "mouse",
		"crow", "fish", "whale", "duck", "swan", "goat", "pig",
		"horse", "sheep", "koala", "zebra", "puma", "crane", "sparrow", "beaver",
		"gopher", "otter", "badger", "pigeon", "lemur", "condor", "vulture", "ibis",
		"camel", "llama", "ferret", "marten", "parrot", "raven", "robin", "squid",
		"crab", "shrimp", "lobster", "salmon", "trout", "bass", "perch", "carp",
	}
)

func generateSessionName() string {
	adj := petnameAdjectives[rand.Intn(len(petnameAdjectives))]
	ani := petnameAnimals[rand.Intn(len(petnameAnimals))]
	return adj + "-" + ani
}

var (
	flagSessionID   = flag.String("session-id", "", "Session ID (auto-generated if empty)")
	flagDatabase    = flag.String("database", "", "Path to SQLite database")
	flagTemperature = flag.Float64("temperature", 0.7, "Model temperature (0-2)")
	flagContextSize = flag.Int("context-size", 25000, "Context window size")
	flagList        = flag.Bool("list-sessions", false, "Show session picker on startup")
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Padding(0, 1)

	userStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Padding(0, 1)

	assistantStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("238")).
			Foreground(lipgloss.Color("213")).
			Bold(true).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("86")).
			Padding(0, 2)

	keymapStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	selectedSessionStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("238")).
				Foreground(lipgloss.Color("86")).
				Padding(0, 2)

	wrapStyle = lipgloss.NewStyle()

	toolStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type sessionItem struct {
	id        int64
	sessionID string
	createdAt string
	msgCount  int
}

func (s sessionItem) Title() string {
	if len(s.sessionID) > 12 {
		return s.sessionID[:12] + "..."
	}
	return s.sessionID
}

func (s sessionItem) Description() string {
	return fmt.Sprintf("%s | %d messages", s.createdAt, s.msgCount)
}

func (s sessionItem) FilterValue() string {
	return s.sessionID
}

type viewMode int

const (
	viewChat viewMode = iota
	viewSessionList
)

type llmResponseMsg struct {
	resp *llm.Message
	err  error
}

type sessionsLoadedMsg struct {
	sessions []sessionItem
	err      error
}

type messagesLoadedMsg struct {
	messages []db.Message
	err      error
}

type searchResultMsg struct {
	content string
	err     error
}

type model struct {
	client      *llm.Generator
	messages    []llm.Message
	input       textinput.Model
	sessionList list.Model
	sessions    []sessionItem
	viewMode    viewMode
	width       int
	height      int
	loading     bool
	useTools    bool
	err         error
	convID      int64
	sessionID   string
	msgCount    int
	temperature float32
	contextSize int
	embedder    *embedder.Embedder
	embedderErr error
	vectorStore *vectorstore.Store
}

func usernameToUID(username string) int64 {
	hash := sha256.Sum256([]byte(username))
	uid, _ := strconv.ParseInt(hex.EncodeToString(hash[:4]), 16, 64)
	if uid < 0 {
		uid = -uid
	}
	return uid
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type a message..."
	ti.Focus()
	ti.CharLimit = 2000
	ti.SetWidth(60)

	client, err := llm.NewClient(context.Background(), -999)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create LLM client: %v\n", err)
		os.Exit(1)
	}
	client.SetContextSize(*flagContextSize)
	client.SetTemperature(float32(*flagTemperature))
	rand.Seed(time.Now().UnixNano())

	var emb *embedder.Embedder
	var embErr error
	var vs *vectorstore.Store
	var convID int64
	if *flagDatabase != "" {
		if err := db.Open(*flagDatabase); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
			os.Exit(1)
		}
		if err := db.Migrate(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to migrate database: %v\n", err)
			os.Exit(1)
		}
		emb, embErr = embedder.New("nomic-embed-text", 768)
		vs = vectorstore.New()
	}

	sessionID := *flagSessionID
	if sessionID == "" {
		sessionID = generateSessionName()
	}

	if *flagDatabase != "" {
		conv, err := db.GetOrCreateSessionBySessionID(context.Background(), sessionID, 1, 1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create session: %v\n", err)
			os.Exit(1)
		}
		convID = conv.ID
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = selectedSessionStyle
	delegate.Styles.SelectedDesc = selectedSessionStyle
	sessionList := list.New([]list.Item{}, delegate, 60, 15)
	sessionList.Title = "Sessions"
	sessionList.SetShowStatusBar(false)
	sessionList.SetShowHelp(false)
	sessionList.SetFilteringEnabled(true)

	m := model{
		client:      client,
		messages:    []llm.Message{{Role: "system", Content: llm.SystemPrompt}},
		input:       ti,
		sessionList: sessionList,
		viewMode:    viewChat,
		loading:     false,
		useTools:    false,
		convID:      convID,
		sessionID:   sessionID,
		msgCount:    0,
		temperature: float32(*flagTemperature),
		contextSize: *flagContextSize,
		embedder:    emb,
		embedderErr: embErr,
		vectorStore: vs,
	}

	if *flagList {
		m.viewMode = viewSessionList
	}

	return m
}

func (m model) Init() tea.Cmd {
	if m.viewMode == viewSessionList {
		return m.loadSessionsCmd()
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.sessionList.SetSize(msg.Width-4, msg.Height-8)
		return m, nil

	case llmResponseMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		if msg.resp != nil {
			m.messages = append(m.messages, *msg.resp)
			if m.convID > 0 {
				if err := db.InsertMessage(context.Background(), &db.Message{
					ConversationID: m.convID,
					Role:           msg.resp.Role,
					Content:        msg.resp.Content,
					TelegramMsgID:  0,
				}); err != nil {
					slog.Error("failed to insert assistant message", "err", err)
					m.err = err
				} else {
					m.msgCount++
					if m.msgCount >= maxMessages {
						if err := db.CompactConversation(context.Background(), m.convID, ""); err != nil {
							slog.Error("failed to compact conversation", "err", err)
						} else {
							m.msgCount = 0
							m.messages = []llm.Message{{Role: "system", Content: llm.SystemPrompt}}
						}
					}
				}
			}
		}
		m.err = nil
		return m, nil

	case sessionsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.sessions = msg.sessions
		items := make([]list.Item, len(m.sessions))
		for i, s := range m.sessions {
			items[i] = s
		}
		m.sessionList.SetItems(items)
		return m, nil

	case messagesLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.messages = []llm.Message{{Role: "system", Content: llm.SystemPrompt}}
		for _, msg := range msg.messages {
			if msg.Role == "system" {
				continue
			}
			m.messages = append(m.messages, llm.Message{
				Role:    msg.Role,
				Content: msg.Content,
			})
			m.msgCount++
		}
		m.viewMode = viewChat
		m.input.Focus()
		return m, nil

	case searchResultMsg:
		if msg.err != nil {
			m.err = msg.err
			m.input.SetValue("")
			return m, nil
		}
		m.messages = append(m.messages, llm.Message{
			Role:    "assistant",
			Content: msg.content,
		})
		m.input.SetValue("")
		return m, nil

	case tea.KeyMsg:
		if m.viewMode == viewSessionList {
			switch msg.String() {
			case "ctrl+c", "esc":
				m.viewMode = viewChat
				return m, nil
			}
		}
	}

	switch m.viewMode {
	case viewChat:
		return m.updateChat(msg)
	case viewSessionList:
		return m.updateSessionList(msg)
	}
	return m, nil
}

func (m model) updateChat(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				return m, nil
			}
			return m.handleSend(text)
		case "ctrl+y":
			m.useTools = !m.useTools
			return m, nil
		case "ctrl+s":
			m.viewMode = viewSessionList
			return m, m.loadSessionsCmd()
		}
	}
	return m, cmd
}

func (m model) updateSessionList(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.sessionList, cmd = m.sessionList.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.sessionList.SelectedItem().(sessionItem); ok {
				m.convID = item.id
				m.sessionID = item.sessionID
				m.msgCount = item.msgCount
				return m, m.loadMessagesCmd(m.convID)
			}
			return m, nil
		case "ctrl+c", "esc":
			m.viewMode = viewChat
			m.input.Focus()
			return m, nil
		}
	}
	return m, cmd
}

func (m model) handleSend(text string) (tea.Model, tea.Cmd) {
	if text == "/exit" || text == "/quit" {
		return m, tea.Quit
	}

	if text == "/reset" {
		m.messages = []llm.Message{{Role: "system", Content: llm.SystemPrompt}}
		m.err = nil
		m.input.SetValue("")
		return m, nil
	}

	if text == "/sessions" {
		m.viewMode = viewSessionList
		m.input.SetValue("")
		return m, m.loadSessionsCmd()
	}

	if strings.HasPrefix(text, "/temp ") || strings.HasPrefix(text, "/temperature ") {
		parts := strings.Fields(text)
		if len(parts) == 2 {
			if temp, err := strconv.ParseFloat(parts[1], 32); err == nil && temp >= 0 && temp <= 2 {
				m.temperature = float32(temp)
				m.client.SetTemperature(m.temperature)
			}
		}
		m.input.SetValue("")
		return m, nil
	}

	if strings.HasPrefix(text, "/search ") {
		query := strings.TrimSpace(text[8:])
		if m.embedder == nil || m.vectorStore == nil {
			if m.embedderErr != nil {
				m.err = fmt.Errorf("embedder initialization failed: %w", m.embedderErr)
			} else {
				m.err = fmt.Errorf("search requires --database flag")
			}
			m.input.SetValue("")
			return m, nil
		}
		return m, m.searchCmd(query)
	}

	if m.convID > 0 {
		if err := db.InsertMessage(context.Background(), &db.Message{
			ConversationID: m.convID,
			Role:           "user",
			Content:        text,
			TelegramMsgID:  0,
		}); err != nil {
			m.err = fmt.Errorf("failed to save message: %w", err)
			m.input.SetValue("")
			return m, nil
		}
		m.msgCount++
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

func (m model) searchCmd(query string) tea.Cmd {
	return func() tea.Msg {
		embeddings, err := m.embedder.Embed(context.Background(), []string{query})
		if err != nil {
			return searchResultMsg{err: err}
		}

		results := m.vectorStore.Search(embeddings[0], 5)
		if len(results) == 0 {
			return searchResultMsg{content: "No results found."}
		}

		var b strings.Builder
		b.WriteString("Found results:\n")
		for i, r := range results {
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, r.Content))
		}
		return searchResultMsg{content: b.String()}
	}
}

func (m model) loadSessionsCmd() tea.Cmd {
	return func() tea.Msg {
		sessions, err := db.GetSessions(context.Background(), 50)
		if err != nil {
			return sessionsLoadedMsg{err: err}
		}
		items := make([]sessionItem, len(sessions))
		for i, s := range sessions {
			items[i] = sessionItem{
				id:        s.ID,
				sessionID: s.SessionID,
				createdAt: s.CreatedAt.Format("2006-01-02 15:04"),
				msgCount:  s.MsgCount,
			}
		}
		return sessionsLoadedMsg{sessions: items}
	}
}

func (m model) loadMessagesCmd(convID int64) tea.Cmd {
	return func() tea.Msg {
		messages, err := db.GetConversationMessages(context.Background(), convID, 100)
		if err != nil {
			return messagesLoadedMsg{err: err}
		}
		return messagesLoadedMsg{messages: messages}
	}
}

func (m model) View() tea.View {
	if m.viewMode == viewSessionList {
		return tea.NewView(m.viewSessionListUI())
	}
	return tea.NewView(m.viewChatUI())
}

func (m model) viewSessionListUI() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("Sessions"))
	b.WriteString("\n\n")
	b.WriteString(m.sessionList.View())
	b.WriteString("\n")
	b.WriteString(keymapStyle.Render("enter: select | esc: back | ctrl+c: quit"))
	return b.String()
}

func (m model) viewChatUI() string {
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
			style = toolStyle
			prefix = "Tool"
		default:
			style = lipgloss.NewStyle()
			prefix = msg.Role
		}

		b.WriteString(style.Render(fmt.Sprintf("[%s]", prefix)))
		b.WriteString("  ")

		wrapWidth := m.width - 10
		if wrapWidth < 20 {
			wrapWidth = 80
		}
		wrapped := wrapStyle.Width(wrapWidth).Render(msg.Content)
		b.WriteString(wrapped)
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

	sessionInfo := fmt.Sprintf("Session: %s | Msgs: %d/%d", m.sessionID, m.msgCount, maxMessages)
	sessionInfo += fmt.Sprintf(" | Temp: %.1f | Ctx: %d", m.temperature, m.contextSize)

	b.WriteString("\n")
	b.WriteString(keymapStyle.Render(sessionInfo))
	b.WriteString("\n")
	b.WriteString(keymapStyle.Render("ctrl+c: quit | ctrl+y: tools | ctrl+s: sessions | /search <q> | /temp <0-2>"))

	return b.String()
}

func main() {
	flag.Parse()

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *flagDatabase != "" {
		db.Close()
	}
}
