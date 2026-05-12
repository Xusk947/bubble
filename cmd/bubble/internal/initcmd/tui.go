package initcmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

type InitResult struct {
	Name        string
	Module      string
	BubbleModule string
	DB          string
	Cache       string
	Queue       string
	S3          bool
}

var errInitCanceled = errors.New("init canceled")

func canUseTUI() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())
}

func runTUI(initial Flags) (InitResult, error) {
	m := newInitModel(initial)
	p := tea.NewProgram(m)
	out, err := p.Run()
	if err != nil {
		return InitResult{}, err
	}
	res, ok := out.(initModel)
	if !ok {
		return InitResult{}, errors.New("unexpected tui model")
	}
	if res.canceled {
		return InitResult{}, errInitCanceled
	}
	return InitResult{
		Name:         strings.TrimSpace(res.name.Value()),
		Module:       strings.TrimSpace(res.module.Value()),
		BubbleModule: strings.TrimSpace(res.bubble.Value()),
		DB:           strings.TrimSpace(res.selectedDB()),
		Cache:        strings.TrimSpace(res.selectedCache()),
		Queue:        strings.TrimSpace(res.selectedQueue()),
		S3:           res.enableS3,
	}, nil
}

type initStep int

const (
	stepName initStep = iota
	stepModule
	stepBubbleModule
	stepDB
	stepCache
	stepQueue
	stepS3
	stepDone
)

type initModel struct {
	step     initStep
	canceled bool

	baseDir string

	name   textinput.Model
	module textinput.Model
	bubble textinput.Model

	dbChoices []string
	dbCursor  int
	dbSelected int

	cacheChoices []string
	cacheCursor  int
	cacheSelected int

	queueChoices []string
	queueCursor  int
	queueSelected int

	s3Cursor int
	enableS3 bool

	err error
}

func newInitModel(initial Flags) initModel {
	name := textinput.New()
	name.Placeholder = "hello-world"
	name.Prompt = "app name: "
	name.SetValue(strings.TrimSpace(initial.Name))
	name.Focus()

	module := textinput.New()
	module.Placeholder = "github.com/user/app"
	module.Prompt = "go module: "
	module.SetValue(strings.TrimSpace(initial.Module))

	bubble := textinput.New()
	bubble.Placeholder = "github.com/user/bubble"
	bubble.Prompt = "bubble module: "
	bubble.SetValue(strings.TrimSpace(initial.BubbleModule))

	dbChoices := []string{dbNone, dbSQLite, dbPostgres}
	dbCursor := 0
	switch normalizeDB(initial.DB) {
	case dbSQLite:
		dbCursor = 1
	case dbPostgres:
		dbCursor = 2
	default:
		dbCursor = 0
	}
	dbSelected := dbCursor

	step := stepName
	if strings.TrimSpace(initial.Name) != "" {
		step = stepModule
		name.Blur()
		module.Focus()
	}
	if strings.TrimSpace(initial.Name) != "" && strings.TrimSpace(initial.Module) != "" {
		inferred := inferBubbleModuleFromAppModule(initial.Module)
		if strings.TrimSpace(bubble.Value()) == "" && isFetchableModule(inferred) {
			bubble.SetValue(inferred)
			step = stepDB
			name.Blur()
			module.Blur()
			bubble.Blur()
		} else {
			step = stepBubbleModule
			name.Blur()
			module.Blur()
			bubble.Focus()
		}
	}

	cacheChoices := []string{cacheLocal, cacheRedis}
	cacheCursor := 0
	if normalizeCache(initial.Cache) == cacheRedis {
		cacheCursor = 1
	}
	cacheSelected := cacheCursor

	queueChoices := []string{queueNone, queueNATS, queueKafka}
	queueCursor := 0
	switch normalizeQueue(initial.Queue) {
	case queueNATS:
		queueCursor = 1
	case queueKafka:
		queueCursor = 2
	default:
		queueCursor = 0
	}
	queueSelected := queueCursor

	return initModel{
		step:      step,
		baseDir:   strings.TrimSpace(initial.Dir),
		name:      name,
		module:    module,
		bubble:    bubble,
		dbChoices: dbChoices,
		dbCursor:  dbCursor,
		dbSelected: dbSelected,
		cacheChoices: cacheChoices,
		cacheCursor:  cacheCursor,
		cacheSelected: cacheSelected,
		queueChoices: queueChoices,
		queueCursor:  queueCursor,
		queueSelected: queueSelected,
		s3Cursor:      0,
		enableS3:      initial.S3,
	}
}

func (m initModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.canceled = true
			return m, tea.Quit
		}
	}

	switch m.step {
	case stepName:
		return m.updateName(msg)
	case stepModule:
		return m.updateModule(msg)
	case stepBubbleModule:
		return m.updateBubbleModule(msg)
	case stepDB:
		return m.updateDB(msg)
	case stepCache:
		return m.updateCache(msg)
	case stepQueue:
		return m.updateQueue(msg)
	case stepS3:
		return m.updateS3(msg)
	default:
		return m, tea.Quit
	}
}

func (m initModel) View() string {
	title := lipgloss.NewStyle().Bold(true).Render("bubble init")
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(helpText(m.step))
	content := ""
	summary := m.renderSummary()

	switch m.step {
	case stepName:
		content = m.name.View()
	case stepModule:
		content = m.module.View()
	case stepBubbleModule:
		content = m.bubble.View()
	case stepDB:
		content = m.renderSingleChoice("db:", m.dbChoices, m.dbCursor)
	case stepCache:
		content = m.renderSingleChoice("cache:", m.cacheChoices, m.cacheCursor)
	case stepQueue:
		content = m.renderSingleChoice("queue:", m.queueChoices, m.queueCursor)
	case stepS3:
		content = m.renderToggle("s3:", []string{"Enable S3 config"}, m.s3Cursor, []bool{m.enableS3})
	default:
		content = ""
	}

	errLine := ""
	if m.err != nil {
		errLine = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(m.err.Error())
	}

	lines := []string{
		title,
		"",
		help,
	}
	if summary != "" {
		lines = append(lines, "", summary)
	}
	lines = append(lines, "", content)
	if errLine != "" {
		lines = append(lines, "", errLine)
	}
	return strings.Join(lines, "\n") + "\n"
}

func (m initModel) updateName(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.name, cmd = m.name.Update(msg)
	m.err = nil

	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "enter" {
		name := strings.TrimSpace(m.name.Value())
		if name == "" {
			m.err = errors.New("app name is empty")
			return m, nil
		}
		if err := validateTargetDir(m.baseDir, name); err != nil {
			m.err = err
			return m, nil
		}
		m.step = stepModule
		m.name.Blur()
		m.module.Focus()
		return m, textinput.Blink
	}
	return m, cmd
}

func (m initModel) updateModule(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.module, cmd = m.module.Update(msg)
	m.err = nil

	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "enter" {
		if strings.TrimSpace(m.module.Value()) == "" {
			m.err = errors.New("go module is empty")
			return m, nil
		}
		inferred := inferBubbleModuleFromAppModule(m.module.Value())
		if strings.TrimSpace(m.bubble.Value()) == "" && isFetchableModule(inferred) {
			m.bubble.SetValue(inferred)
			m.step = stepDB
			m.module.Blur()
			m.bubble.Blur()
			return m, nil
		}
		m.step = stepBubbleModule
		m.module.Blur()
		m.bubble.Focus()
		return m, nil
	}
	return m, cmd
}

func (m initModel) updateBubbleModule(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.bubble, cmd = m.bubble.Update(msg)
	m.err = nil

	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "enter" {
		if strings.TrimSpace(m.bubble.Value()) == "" {
			m.err = errors.New("bubble module is empty")
			return m, nil
		}
		if !isFetchableModule(m.bubble.Value()) {
			m.err = errors.New("bubble module is invalid (expected github.com/...)")
			return m, nil
		}
		m.step = stepDB
		m.bubble.Blur()
		return m, nil
	}
	return m, cmd
}

func (m initModel) updateDB(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.err = nil
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "up", "k":
			if m.dbCursor > 0 {
				m.dbCursor--
			}
		case "down", "j":
			if m.dbCursor < len(m.dbChoices)-1 {
				m.dbCursor++
			}
		case "enter":
			m.dbSelected = m.dbCursor
			db := m.selectedDB()
			if err := validateDB(db); err != nil {
				m.err = err
				return m, nil
			}
			m.step = stepCache
			return m, nil
		}
	}
	return m, nil
}

func (m initModel) selectedDB() string {
	if m.dbSelected < 0 || m.dbSelected >= len(m.dbChoices) {
		return dbNone
	}
	return m.dbChoices[m.dbSelected]
}

func (m initModel) renderDB() string {
	return m.renderSingleChoice("db:", m.dbChoices, m.dbCursor)
}

func (m initModel) updateCache(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.err = nil
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "up", "k":
			if m.cacheCursor > 0 {
				m.cacheCursor--
			}
		case "down", "j":
			if m.cacheCursor < len(m.cacheChoices)-1 {
				m.cacheCursor++
			}
		case "enter":
			m.cacheSelected = m.cacheCursor
			if err := validateCache(m.selectedCache()); err != nil {
				m.err = err
				return m, nil
			}
			m.step = stepQueue
			return m, nil
		}
	}
	return m, nil
}

func (m initModel) selectedCache() string {
	if m.cacheSelected < 0 || m.cacheSelected >= len(m.cacheChoices) {
		return cacheLocal
	}
	return m.cacheChoices[m.cacheSelected]
}

func (m initModel) updateQueue(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.err = nil
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "up", "k":
			if m.queueCursor > 0 {
				m.queueCursor--
			}
		case "down", "j":
			if m.queueCursor < len(m.queueChoices)-1 {
				m.queueCursor++
			}
		case "enter":
			m.queueSelected = m.queueCursor
			if err := validateQueue(m.selectedQueue()); err != nil {
				m.err = err
				return m, nil
			}
			m.step = stepS3
			return m, nil
		}
	}
	return m, nil
}

func (m initModel) selectedQueue() string {
	if m.queueSelected < 0 || m.queueSelected >= len(m.queueChoices) {
		return queueNone
	}
	return m.queueChoices[m.queueSelected]
}

func (m initModel) updateS3(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.err = nil
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "up", "k":
			if m.s3Cursor > 0 {
				m.s3Cursor--
			}
		case "down", "j":
			if m.s3Cursor < 0 {
				m.s3Cursor = 0
			}
		case " ":
			m.enableS3 = !m.enableS3
		case "enter":
			m.step = stepDone
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m initModel) renderSummary() string {
	lines := make([]string, 0, 7)
	if m.step > stepName && strings.TrimSpace(m.name.Value()) != "" {
		lines = append(lines, "app: "+strings.TrimSpace(m.name.Value()))
	}
	if m.step > stepModule && strings.TrimSpace(m.module.Value()) != "" {
		lines = append(lines, "module: "+strings.TrimSpace(m.module.Value()))
	}
	if m.step > stepBubbleModule && strings.TrimSpace(m.bubble.Value()) != "" {
		lines = append(lines, "bubble: "+strings.TrimSpace(m.bubble.Value()))
	}
	if m.step > stepBubbleModule {
		lines = append(lines, "db: "+m.selectedDB())
	}
	if m.step > stepDB {
		lines = append(lines, "cache: "+m.selectedCache())
	}
	if m.step > stepCache {
		lines = append(lines, "queue: "+m.selectedQueue())
	}
	if m.step > stepQueue {
		s3 := "off"
		if m.enableS3 {
			s3 = "on"
		}
		lines = append(lines, "s3: "+s3)
	}
	return strings.Join(lines, "\n")
}

func (m initModel) renderSingleChoice(title string, choices []string, cursor int) string {
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))

	lines := make([]string, 0, 1+len(choices))
	lines = append(lines, title)
	for i, c := range choices {
		prefix := "  "
		if i == cursor {
			prefix = cursorStyle.Render("› ")
		}
		style := normalStyle
		if i == cursor {
			style = selectedStyle
		}
		lines = append(lines, fmt.Sprintf("%s%s", prefix, style.Render(c)))
	}
	return strings.Join(lines, "\n")
}

func (m initModel) renderToggle(title string, choices []string, cursor int, enabled []bool) string {
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	checkedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))

	lines := make([]string, 0, 1+len(choices))
	lines = append(lines, title)
	for i, c := range choices {
		prefix := "  "
		if i == cursor {
			prefix = cursorStyle.Render("› ")
		}

		on := false
		if i >= 0 && i < len(enabled) {
			on = enabled[i]
		}

		box := "□"
		style := normalStyle
		if on {
			box = "■"
			style = checkedStyle
		}

		lines = append(lines, fmt.Sprintf("%s%s %s", prefix, style.Render(box), style.Render(c)))
	}
	return strings.Join(lines, "\n")
}

func helpText(step initStep) string {
	switch step {
	case stepDB, stepCache, stepQueue:
		return "↑/↓ to move • enter to select • esc to cancel"
	case stepS3:
		return "↑/↓ to move • space to toggle • enter to continue • esc to cancel"
	default:
		return "enter to continue • esc to cancel"
	}
}

func validateTargetDir(baseDir string, appName string) error {
	target := strings.TrimSpace(baseDir)
	if target == "" {
		target = "./" + strings.TrimSpace(appName)
	}

	stat, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if stat.IsDir() {
		return errors.New("target directory already exists")
	}
	return errors.New("target path exists and is not a directory")
}
