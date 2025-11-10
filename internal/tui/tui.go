package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"tree/internal/config"
	"tree/internal/tree"

	"tree/internal/types"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

// DirEntry — элемент списка для Bubble Tea
type DirEntry struct {
	types.Entry
	path string
}

func (d DirEntry) Title() string {
	if d.Depth == 0 {
		return filepath.Base(d.path) + "/"
	}

	// Символы для рисования дерева
	const (
		vLine   = "│"
		hLine   = "─"
		cornerR = "├"
		cornerB = "└"
		space   = " "
	)

	prefix := ""
	for i := 1; i < d.Depth; i++ {
		prefix += vLine + space + space + space
	}

	// Определяем, является ли элемент последним в своей директории
	// Это сложно без хранения структуры дерева — пока упростим:
	// Всегда используем "├──", кроме последнего элемента (если знаете его индекс)
	// Для простоты — просто добавим "├──" на всех уровнях, кроме корня

	prefix += cornerR + hLine + hLine + hLine

	name := filepath.Base(d.path)
	if d.Info.IsDir() {
		name += "/"
	}

	return prefix + name
}

func (d DirEntry) Description() string {
	if d.Info.IsDir() {
		return "directory"
	}
	return fmt.Sprintf("%d bytes", d.Info.Size())
}

func (d DirEntry) FilterValue() string { return d.path }

// Model — основная модель TUI
type Model struct {
	ctx          context.Context
	cfg          *config.Config
	rootPath     string
	entries      []list.Item
	list         list.Model
	viewport     viewport.Model
	showFileView bool
	err          error
}

// NewModel создаёт новую модель TUI
func NewModel(ctx context.Context, cfg *config.Config, rootPath string) (Model, error) {
	// Сканируем директорию
	walkResult, err := tree.WalkDirWithContext(ctx, rootPath, cfg, false)
	if err != nil {
		return Model{}, err
	}

	// Преобразуем записи
	var items []list.Item
	for _, entry := range walkResult.Entries {
		fullPath := entry.Path
		if entry.Depth == 0 {
			fullPath = rootPath
		} else {
			fullPath = filepath.Join(rootPath, entry.Path)
		}

		// Создаём DirEntry и добавляем в список
		item := DirEntry{
			Entry: entry,
			path:  fullPath,
		}
		items = append(items, item) // ← Важно: добавляем именно DirEntry, а не types.Entry
	}

	// Создаём список
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = fmt.Sprintf("📁 %s", rootPath)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")).MarginLeft(2)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	// Получаем размер терминала
	width, height, _ := term.GetSize(os.Stdout.Fd())
	if width > 0 && height > 0 {
		l.SetSize(width, height-5) // -5 для заголовка и статус-бара
	}
	// Viewport для просмотра файлов
	vp := viewport.New(80, 20)

	return Model{
		ctx:      ctx,
		cfg:      cfg,
		rootPath: rootPath,
		entries:  items,
		list:     l,
		viewport: vp,
	}, nil
}

// Init инициализация модели
func (m Model) Init() tea.Cmd {
	return nil
}

// Update обработка событий
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch key := msg.String(); key {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			if !m.showFileView {
				item, ok := m.list.SelectedItem().(DirEntry)
				if ok {
					if item.Info.IsDir() {
						// Рекурсивно открываем поддиректорию
						newModel, err := NewModel(m.ctx, m.cfg, item.path)
						if err != nil {
							m.err = err
							return m, tea.Quit
						}
						return newModel, nil
					} else {
						// Показываем содержимое файла
						content, err := os.ReadFile(item.path)
						if err != nil {
							m.err = err
							return m, nil
						}
						m.viewport.SetContent(string(content))
						m.showFileView = true
					}
				}
			} else {
				m.showFileView = false
			}

		case "esc", "backspace":
			if m.showFileView {
				m.showFileView = false
			} else {
				// Возвращаемся на уровень выше
				parent := filepath.Dir(m.rootPath)
				if parent != m.rootPath {
					newModel, err := NewModel(m.ctx, m.cfg, parent)
					if err != nil {
						m.err = err
						return m, tea.Quit
					}
					return newModel, nil
				}
			}
		}
	}

	if !m.showFileView {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View рендеринг интерфейса
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress any key to exit", m.err)
	}

	if m.showFileView {
		return lipgloss.JoinVertical(lipgloss.Top,
			lipgloss.NewStyle().Padding(1).Render("📄 File Viewer (ESC to go back)"),
			m.viewport.View(),
		)
	}

	// Добавляем заголовок с информацией
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		Render(fmt.Sprintf("📁 %s — %d items", m.rootPath, len(m.entries)))

	return lipgloss.JoinVertical(lipgloss.Top,
		header,
		m.list.View(),
	)
}

// Run запускает TUI
func Run(ctx context.Context, cfg *config.Config, rootPath string) error {
	model, err := NewModel(ctx, cfg, rootPath)
	if err != nil {
		return err
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
