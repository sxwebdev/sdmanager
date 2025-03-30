package sdmanager

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Создать пункты главного меню
func GetMenuItems() []list.Item {
	items := []list.Item{
		MenuItem{Title: string(ActionStartService), Action: ActionStartService},
		MenuItem{Title: string(ActionStopService), Action: ActionStopService},
		MenuItem{Title: string(ActionRestartService), Action: ActionRestartService},
		MenuItem{Title: string(ActionViewLogs), Action: ActionViewLogs},
		MenuItem{Title: string(ActionInstallService), Action: ActionInstallService},
		MenuItem{Title: string(ActionExit), Action: ActionExit},
	}
	return items
}

// Создать модель меню
func NewMenuModel() MenuModel {
	const defaultWidth = 40

	l := list.New(GetMenuItems(), ItemDelegate{}, defaultWidth, ListHeight)
	l.Title = "Systemd Manager"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle
	l.Styles.PaginationStyle = PaginationStyle
	l.Styles.HelpStyle = HelpStyle

	return MenuModel{
		List:     l,
		Choice:   "",
		Quitting: false,
	}
}

// Обработка событий в меню
func UpdateMenu(msg tea.Msg, model MenuModel) (MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.List.SetWidth(msg.Width)
		return model, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			model.Quitting = true
			return model, tea.Quit

		case "enter":
			item, ok := model.List.SelectedItem().(MenuItem)
			if ok {
				model.Choice = item.Action
			}
			return model, nil
		}
	}

	var cmd tea.Cmd
	model.List, cmd = model.List.Update(msg)
	return model, cmd
}

// Отрисовка меню
func ViewMenu(model MenuModel) string {
	return "\n" + model.List.View()
}
