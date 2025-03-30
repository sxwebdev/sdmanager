package sdmanager

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Создание новой модели приложения
func NewApplication(o AppOptions) AppModel {
	if o.ctx == nil {
		o.ctx = context.Background()
	}

	return AppModel{
		Mode:       ModeMainMenu,
		MenuModel:  NewMenuModel(),
		Message:    "",
		Error:      "",
		FatalError: false,

		options: o,
	}
}

// Инициализация приложения
func (m AppModel) Init() tea.Cmd {
	return nil
}

// Обновление состояния приложения
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Если есть фатальная ошибка, сразу выходим
	if m.FatalError {
		return m, tea.Quit
	}

	switch m.Mode {
	case ModeMainMenu:
		// Обновляем модель меню
		menuModel, cmd := UpdateMenu(msg, m.MenuModel)
		m.MenuModel = menuModel

		// Проверяем выбор пользователя
		if m.MenuModel.Choice != "" {
			switch m.MenuModel.Choice {
			case ActionInstallService:
				// Переключаемся в режим установки сервиса
				m.Mode = ModeInstallService
				m.InstallModel = NewInstallModel(m.options)
				return m, nil

			case ActionStartService:
				// Переходим к вводу имени сервиса для запуска
				m.Mode = ModeServiceInput
				m.ServiceInputModel = NewServiceInputModel(ActionStart)
				return m, nil

			case ActionStopService:
				// Переходим к вводу имени сервиса для остановки
				m.Mode = ModeServiceInput
				m.ServiceInputModel = NewServiceInputModel(ActionStop)
				return m, nil

			case ActionRestartService:
				// Переходим к вводу имени сервиса для перезапуска
				m.Mode = ModeServiceInput
				m.ServiceInputModel = NewServiceInputModel(ActionRestart)
				return m, nil

			case ActionViewLogs:
				// Переходим к вводу имени сервиса для просмотра логов
				m.Mode = ModeServiceInput
				m.ServiceInputModel = NewServiceInputModel(ActionViewLog)
				return m, nil

			case ActionExit:
				// Выход из программы
				return m, tea.Quit
			}
		}

		// Если выход из меню, но не выбрано действие
		if m.MenuModel.Quitting {
			return m, tea.Quit
		}

		return m, cmd

	case ModeServiceInput:
		// Обновляем модель ввода имени сервиса
		serviceModel, cmd, err := UpdateServiceInput(m.options.ctx, msg, m.ServiceInputModel)
		m.ServiceInputModel = serviceModel

		// Обрабатываем ошибку
		if err != nil {
			m.Error = err.Error()
			m.FatalError = true
			return m, tea.Quit
		}

		// Если ввод имени сервиса завершен
		if m.ServiceInputModel.Quitting {
			// Если есть результат операции, выводим его и выходим
			if m.ServiceInputModel.ResultMsg != "" {
				fmt.Println(m.ServiceInputModel.ResultMsg)
			}

			// В любом случае выходим
			return m, tea.Quit
		}

		return m, cmd

	case ModeInstallService:
		// Обновляем модель установки
		installModel, cmd, err := UpdateInstall(msg, m.InstallModel)
		m.InstallModel = installModel

		// Обрабатываем ошибку
		if err != nil {
			m.Error = err.Error()
			m.FatalError = true
			return m, tea.Quit
		}

		// Если установка завершена
		if m.InstallModel.Quitting {
			// Если есть результат операции, выводим его и выходим
			if m.InstallModel.ResultMsg != "" {
				fmt.Println(m.InstallModel.ResultMsg)
			} else if !m.InstallModel.Aborted {
				fmt.Println("Установка успешно завершена")
			} else {
				fmt.Println("Установка прервана")
			}

			// В любом случае выходим
			return m, tea.Quit
		}

		return m, cmd

	case ModeError:
		// В режиме ошибки просто выходим при любом действии
		return m, tea.Quit
	}

	return m, nil
}

// Отображение интерфейса приложения
func (m AppModel) View() string {
	// При фатальной ошибке показываем сообщение об ошибке
	if m.FatalError {
		return FormatError(m.Error)
	}

	switch m.Mode {
	case ModeMainMenu:
		var s strings.Builder

		// Отображаем сообщение, если оно есть
		if m.Message != "" {
			s.WriteString(FormatInfo(m.Message) + "\n\n")
		}

		// Отображаем ошибку, если она есть
		if m.Error != "" {
			s.WriteString(FormatError(m.Error) + "\n\n")
		}

		// Отображаем меню
		s.WriteString(ViewMenu(m.MenuModel))

		return s.String()

	case ModeServiceInput:
		return ViewServiceInput(m.ServiceInputModel)

	case ModeInstallService:
		return ViewInstall(m.InstallModel)

	case ModeError:
		return FormatError(m.Error)
	}

	return ""
}

// RunSystemdManager запускает интерфейс управления systemd
func RunSystemdManager(otps ...AppOption) error {
	o := AppOptions{}
	for _, opt := range otps {
		opt(&o)
	}

	appModel := NewApplication(o)
	p := tea.NewProgram(appModel)
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
