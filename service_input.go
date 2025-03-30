package sdmanager

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Инициализация модели ввода имени сервиса
func NewServiceInputModel(action string) ServiceInputModel {
	ti := textinput.New()
	ti.Placeholder = "myservice"
	ti.Focus()
	ti.CharLimit = 255
	ti.Width = 80

	var message string
	switch action {
	case ActionStart:
		message = "Введите имя сервиса для запуска:"
	case ActionStop:
		message = "Введите имя сервиса для остановки:"
	case ActionRestart:
		message = "Введите имя сервиса для перезапуска:"
	case ActionViewLog:
		message = "Введите имя сервиса для просмотра логов:"
	default:
		message = "Введите имя сервиса:"
	}

	return ServiceInputModel{
		Input:     ti,
		Action:    action,
		Message:   message,
		Error:     "",
		ResultMsg: "",
		Quitting:  false,
	}
}

// Обработка событий при вводе имени сервиса
func UpdateServiceInput(msg tea.Msg, model ServiceInputModel) (ServiceInputModel, tea.Cmd, error) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:

			// Если есть предыдущая ошибка, просто очищаем её
			if model.Error != "" {
				model.Error = ""
				return model, nil, nil
			}

			serviceName := model.Input.Value()
			if serviceName == "" {
				model.Error = "Имя сервиса не может быть пустым"
				return model, nil, nil
			}

			// Проверка валидности имени сервиса
			if err := IsValidServiceName(serviceName); err != nil {
				model.Error = err.Error()
				return model, nil, nil
			}

			// Выполняем выбранное действие
			var result string
			var err error
			switch model.Action {
			case ActionStart:
				result, err = StartService(serviceName)
			case ActionStop:
				result, err = StopService(serviceName)
			case ActionRestart:
				result, err = RestartService(serviceName)
			case ActionViewLog:
				result, err = ViewServiceLogs(serviceName)
			}

			if err != nil {
				model.Error = err.Error()
				return model, nil, nil
			}

			model.ResultMsg = result
			model.Quitting = true
			return model, tea.Quit, nil

		case tea.KeyTab:
			// Автозаполнение из placeholder если поле пустое
			if model.Input.Value() == "" && model.Input.Placeholder != "" {
				model.Input.SetValue(model.Input.Placeholder)
			}

		case tea.KeyCtrlC, tea.KeyEsc:
			model.Quitting = true
			return model, tea.Quit, nil
		}
	}

	// Обновляем текстовый ввод
	var cmd tea.Cmd
	model.Input, cmd = model.Input.Update(msg)
	return model, cmd, nil
}

// Отрисовка формы ввода имени сервиса
func ViewServiceInput(model ServiceInputModel) string {
	var s strings.Builder

	s.WriteString(model.Message + "\n\n")
	s.WriteString(model.Input.View() + "\n\n")

	// Отображаем ошибку на отдельной строке, если она есть
	if model.Error != "" {
		s.WriteString(FormatError(model.Error) + "\n\n")
	}

	s.WriteString("Нажмите Enter для подтверждения, Esc (Ctrl+C) для отмены\n")

	return s.String()
}
