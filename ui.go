package sdmanager

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Константы для стилей UI
const (
	ListHeight = 14
)

// Стили для UI
var (
	TitleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	ItemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	SelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	PaginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	HelpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	QuitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	InfoStyle         = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("86"))
	ErrorStyle        = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("196"))
	ViewportStyle     = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(1).BorderForeground(lipgloss.Color("62"))
)

// Делегат для отображения пунктов меню
type ItemDelegate struct{}

func (d ItemDelegate) Height() int                             { return 1 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(MenuItem)
	if !ok {
		return
	}

	str := i.Title
	fn := ItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return SelectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

// Функция для форматирования сообщений об ошибках
func FormatError(err string) string {
	return ErrorStyle.Render("Ошибка: " + err)
}

// Функция для форматирования информационных сообщений
func FormatInfo(info string) string {
	return InfoStyle.Render(info)
}

// Отобразить список опций с выбором
func RenderOptionsList(options []Option, currentOption int) string {
	var sb strings.Builder

	for i, opt := range options {
		checkbox := "[ ]"
		if opt.Selected {
			checkbox = "[x]"
		}

		line := fmt.Sprintf("%s %s", checkbox, opt.Name)

		if i == currentOption {
			sb.WriteString(SelectedItemStyle.Render("> "+line) + "\n")
		} else {
			sb.WriteString("    " + line + "\n")
		}
	}

	sb.WriteString("\nИспользуйте стрелки ↑/↓ для навигации, пробел для выбора, Enter для подтверждения\n")
	return sb.String()
}

// Отобразить выбранные опции
func RenderSelectedOptions(actions UserActions) string {
	var sb strings.Builder

	sb.WriteString("Выбранные опции:\n")
	if actions.ReloadDaemon {
		sb.WriteString("✓ Перезагрузить systemd daemon\n")
	}
	if actions.EnableService {
		sb.WriteString("✓ Активировать (enable) сервис\n")
	}
	if actions.StartService {
		sb.WriteString("✓ Запустить (start) сервис\n")
	}

	return sb.String()
}
