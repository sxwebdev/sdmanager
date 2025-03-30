package sdmanager

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Инициализация модели установки
func NewInstallModel(appOptions AppOptions) InstallModel {
	serviceName := appOptions.serviceName
	if serviceName == "" {
		serviceName = "myservice"
	}

	ti := textinput.New()
	ti.Placeholder = serviceName
	ti.Focus()
	ti.CharLimit = 255
	ti.Width = 80

	currentDir := GetCurrentDir()
	execPath := GetCurrentExecutable()

	// Viewport для предпросмотра unit файла - увеличен размер по высоте
	vp := viewport.New(78, 30)
	vp.Style = ViewportStyle

	return InstallModel{
		State: StateServiceName,
		Config: ServiceConfig{
			ServiceName:      serviceName,
			WorkingDirectory: currentDir,
			ExecStart:        execPath,
			MemoryHigh:       0,
			MemoryMax:        0,
			UnitFilePath:     "/etc/systemd/system",
		},
		Actions: UserActions{
			Overwrite:     false,
			ReloadDaemon:  true, // По умолчанию включены все опции
			EnableService: true,
			StartService:  true,
		},
		Input:          ti,
		Viewport:       vp,
		Message:        fmt.Sprintf("Введите название сервиса (по умолчанию: %s):", serviceName),
		ErrorMsg:       "",
		PreviewContent: "",
		ShowHelp:       true,
		Quitting:       false,
		Aborted:        false,
		ResultMsg:      "",
		Options: []Option{
			{Name: "Перезагрузить systemd daemon", Selected: true},
			{Name: "Активировать (enable) сервис", Selected: true},
			{Name: "Запустить (start) сервис", Selected: true},
		},
		CurrentOption: 0,
	}
}

// Обработка события ввода имени сервиса
func HandleServiceNameInput(model InstallModel, input string) (InstallModel, error) {
	if input == "" {
		input = model.Config.ServiceName // Используем значение по умолчанию
	}

	if err := IsValidServiceName(input); err != nil {
		model.ErrorMsg = err.Error()
		return model, nil
	}

	model.Config.ServiceName = input
	model.State = StateUserName
	model.Message = "Введите имя юзера (оционально):"
	model.Input.SetValue("")
	model.Input.Placeholder = model.Config.UserName

	return model, nil
}

// Обработка события ввода юзера
func HandleUserNameInput(model InstallModel, input string) (InstallModel, error) {
	if err := IsValidUserName(input); err != nil {
		model.ErrorMsg = err.Error()
		return model, nil
	}

	model.Config.UserName = input
	model.State = StateWorkingDirectory
	model.Message = "Введите рабочую директорию сервиса (по умолчанию: текущая директория):"
	model.Input.SetValue("")
	model.Input.Placeholder = model.Config.WorkingDirectory

	return model, nil
}

// Обработка события ввода рабочей директории
func HandleWorkingDirectoryInput(model InstallModel, input string) (InstallModel, error) {
	if input != "" {
		if err := IsValidPath(input); err != nil {
			model.ErrorMsg = err.Error()
			return model, nil
		}
		model.Config.WorkingDirectory = input
	}
	// Если ввод пустой, оставляем текущую директорию по умолчанию

	// Переход к вводу команды запуска
	model.State = StateExecStart
	model.Message = "Введите команду ExecStart (по умолчанию: текущий исполняемый файл):"
	model.Input.SetValue("")
	model.Input.Placeholder = model.Config.ExecStart

	return model, nil
}

// Обработка события ввода команды запуска
func HandleExecStartInput(model InstallModel, input string) (InstallModel, error) {
	if input != "" {
		model.Config.ExecStart = input
	}
	// Если ввод пустой, оставляем значение по умолчанию

	// Переход к вводу StandardOutput
	model.State = StateStandardOutput
	model.Message = "Введите значение для StandardOutput (опционально):"
	model.Input.SetValue("")
	model.Input.Placeholder = ""

	return model, nil
}

// Обработка события ввода StandardOutput
func HandleStandardOutputInput(model InstallModel, input string) (InstallModel, error) {
	model.Config.StandardOutput = input

	// Переход к вводу StandardError
	model.State = StateStandardError
	model.Message = "Введите значение для StandardError (опционально):"
	model.Input.SetValue("")
	model.Input.Placeholder = ""

	return model, nil
}

// Обработка события ввода StandardError
func HandleStandardErrorInput(model InstallModel, input string) (InstallModel, error) {
	model.Config.StandardError = input

	// Переход к вводу SyslogIdentifier
	model.State = StateSyslogIdentifier
	model.Message = "Введите SyslogIdentifier (опционально):"
	model.Input.SetValue("")
	model.Input.Placeholder = ""

	return model, nil
}

// Обработка события ввода SyslogIdentifier
func HandleSyslogIdentifierInput(model InstallModel, input string) (InstallModel, error) {
	model.Config.SyslogIdentifier = input

	// Переход к вводу MemoryHigh
	model.State = StateMemoryHigh
	model.Message = "Введите ограничение MemoryHigh в МБ (0 - не использовать):"
	model.Input.SetValue("")
	model.Input.Placeholder = "0"

	return model, nil
}

// Обработка события ввода MemoryHigh
func HandleMemoryHighInput(model InstallModel, input string) (InstallModel, error) {
	val, err := ParseIntValue(input, 0)
	if err != nil {
		model.ErrorMsg = err.Error()
		return model, nil
	}

	model.Config.MemoryHigh = val

	// Переход к вводу MemoryMax
	model.State = StateMemoryMax
	model.Message = "Введите ограничение MemoryMax в МБ (0 - не использовать):"
	model.Input.SetValue("")
	model.Input.Placeholder = "0"

	return model, nil
}

// Обработка события ввода MemoryMax
func HandleMemoryMaxInput(model InstallModel, input string) (InstallModel, error) {
	val, err := ParseIntValue(input, 0)
	if err != nil {
		model.ErrorMsg = err.Error()
		return model, nil
	}

	model.Config.MemoryMax = val

	// Проверка логики ограничений памяти
	if model.Config.MemoryHigh > 0 && model.Config.MemoryMax > 0 &&
		model.Config.MemoryHigh >= model.Config.MemoryMax {
		model.ErrorMsg = "Ошибка: MemoryHigh должен быть меньше MemoryMax"
		return model, nil
	}

	// Переход к вводу пути unit-файла
	model.State = StateUnitLocation
	model.Message = "Введите путь для сохранения unit-файла (по умолчанию: /etc/systemd/system):"
	model.Input.SetValue("")
	model.Input.Placeholder = "/etc/systemd/system"

	return model, nil
}

// Обработка события ввода пути unit-файла
func HandleUnitLocationInput(model InstallModel, input string) (InstallModel, error) {
	if input == "" {
		model.Config.UnitFilePath = "/etc/systemd/system"
	} else {
		if err := IsValidPath(input); err != nil {
			model.ErrorMsg = err.Error()
			return model, nil
		}
		model.Config.UnitFilePath = input
	}

	// Проверка существования файла
	unitFilePath := filepath.Join(model.Config.UnitFilePath, model.Config.ServiceName+".service")
	if FileExists(unitFilePath) {
		model.State = StateOverwrite
		model.Message = fmt.Sprintf("Файл %s уже существует. Перезаписать? (y/n):", unitFilePath)
		model.Input.SetValue("")
	} else {
		// Переходим к выбору опций
		model.State = StateOptionsSelect
		model.Message = "Выберите опции (пробел для переключения, Enter для подтверждения):"
		model.Input.SetValue("")
	}

	return model, nil
}

// Обработка события ответа на вопрос о перезаписи
func HandleOverwriteInput(model InstallModel, input string) (InstallModel, error) {
	if len(input) > 0 && (input[0] == 'y' || input[0] == 'Y') {
		model.Actions.Overwrite = true
		model.State = StateOptionsSelect
		model.Message = "Выберите опции (пробел для переключения, Enter для подтверждения):"
		model.Input.SetValue("")
	} else {
		model.Aborted = true
		model.Message = "Операция прервана пользователем."
		return model, errors.New("операция прервана пользователем")
	}

	return model, nil
}

// Обработка события выбора опций
func HandleOptionsSelect(model InstallModel) (InstallModel, error) {
	// Применяем выбранные опции
	model.Actions.ReloadDaemon = model.Options[0].Selected
	model.Actions.EnableService = model.Options[1].Selected
	model.Actions.StartService = model.Options[2].Selected

	// Генерируем предпросмотр
	preview, err := GenerateUnitPreview(model.Config)
	if err != nil {
		model.ErrorMsg = fmt.Sprintf("Ошибка при генерации предпросмотра: %s", err)
		return model, nil
	}

	model.PreviewContent = preview
	model.Viewport.SetContent(preview)
	model.State = StatePreviewUnit
	model.Message = "Предпросмотр unit-файла (Enter - сохранить, Esc - отменить):"

	return model, nil
}

// Обработка события подтверждения установки в режиме предпросмотра
func HandlePreviewConfirmation(model InstallModel) (InstallModel, error) {
	// Если уже есть ошибка, очищаем её
	if model.ErrorMsg != "" {
		model.ErrorMsg = ""
		return model, nil
	}

	// Выполняем установку сервиса
	result, err := InstallService(model.Config, model.Actions)
	if err != nil {
		model.ErrorMsg = err.Error()
		return model, nil
	}

	model.ResultMsg = result
	model.State = StateDone
	model.Quitting = true

	return model, nil
}

// Обработка сообщений для установки сервиса
func UpdateInstall(msg tea.Msg, model InstallModel) (InstallModel, tea.Cmd, error) {
	var err error

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// Сначала проверяем, есть ли ошибка - если да, просто сбрасываем её
			if model.ErrorMsg != "" {
				model.ErrorMsg = ""
				return model, nil, nil
			}

			// Обработка Enter в зависимости от текущего состояния
			switch model.State {
			case StateServiceName:
				model, err = HandleServiceNameInput(model, model.Input.Value())
			case StateUserName:
				model, err = HandleUserNameInput(model, model.Input.Value())
			case StateWorkingDirectory:
				model, err = HandleWorkingDirectoryInput(model, model.Input.Value())
			case StateExecStart:
				model, err = HandleExecStartInput(model, model.Input.Value())
			case StateStandardOutput:
				model, err = HandleStandardOutputInput(model, model.Input.Value())
			case StateStandardError:
				model, err = HandleStandardErrorInput(model, model.Input.Value())
			case StateSyslogIdentifier:
				model, err = HandleSyslogIdentifierInput(model, model.Input.Value())
			case StateMemoryHigh:
				model, err = HandleMemoryHighInput(model, model.Input.Value())
			case StateMemoryMax:
				model, err = HandleMemoryMaxInput(model, model.Input.Value())
			case StateUnitLocation:
				model, err = HandleUnitLocationInput(model, model.Input.Value())
			case StateOverwrite:
				model, err = HandleOverwriteInput(model, model.Input.Value())
			case StateOptionsSelect:
				model, err = HandleOptionsSelect(model)
				if err == nil {
					return model, tea.ClearScreen, nil
				}
			case StatePreviewUnit:
				model, err = HandlePreviewConfirmation(model)
				if err == nil && model.Quitting {
					// Успешно завершено, выходим
					return model, tea.Quit, nil
				}
			}

			if err != nil {
				model.ErrorMsg = err.Error()
				return model, nil, nil
			}

		case tea.KeyCtrlC, tea.KeyEsc:
			model.Aborted = true
			model.Message = "Операция прервана пользователем."
			return model, tea.Quit, nil

		case tea.KeyCtrlH:
			// Переключение отображения справки
			model.ShowHelp = !model.ShowHelp

		// Управление выбором опций
		case tea.KeySpace:
			if model.State == StateOptionsSelect {
				// Переключаем выбранную опцию
				model.Options[model.CurrentOption].Selected = !model.Options[model.CurrentOption].Selected
			}

		case tea.KeyTab:
			// Автозаполнение из placeholder если поле пустое
			if model.Input.Value() == "" && model.Input.Placeholder != "" {
				model.Input.SetValue(model.Input.Placeholder)
			}

		case tea.KeyDown:
			if model.State == StateOptionsSelect {
				// Переход к следующей опции
				model.CurrentOption = (model.CurrentOption + 1) % len(model.Options)
			} else if model.State == StatePreviewUnit {
				model.Viewport.LineDown(1)
			}

		case tea.KeyUp:
			if model.State == StateOptionsSelect {
				// Переход к предыдущей опции
				model.CurrentOption = (model.CurrentOption - 1 + len(model.Options)) % len(model.Options)
			} else if model.State == StatePreviewUnit {
				model.Viewport.LineUp(1)
			}

		case tea.KeyPgDown:
			if model.State == StatePreviewUnit {
				model.Viewport.LineDown(10)
			}

		case tea.KeyPgUp:
			if model.State == StatePreviewUnit {
				model.Viewport.LineUp(10)
			}
		}

	case tea.WindowSizeMsg:
		// Обновляем размер viewport при изменении размера окна
		if model.State == StatePreviewUnit {
			model.Viewport.Width = msg.Width - 4
			model.Viewport.Height = msg.Height - 10
		}
	}

	// Обновляем текстовый ввод
	var cmd tea.Cmd
	model.Input, cmd = model.Input.Update(msg)

	return model, cmd, nil
}

// Отрисовка интерфейса установки
func ViewInstall(model InstallModel) string {
	if model.Quitting {
		if model.ResultMsg != "" {
			return model.ResultMsg
		}
		return model.Message + "\n"
	}

	var s strings.Builder

	// Отображаем основное сообщение
	s.WriteString(model.Message + "\n\n")

	// В режиме выбора опций показываем список опций
	if model.State == StateOptionsSelect {
		s.WriteString(RenderOptionsList(model.Options, model.CurrentOption))
	} else if model.State == StatePreviewUnit {
		// В режиме предпросмотра показываем viewport
		s.WriteString(model.Viewport.View() + "\n\n")

		// Показываем выбранные опции
		s.WriteString(RenderSelectedOptions(model.Actions) + "\n")

		s.WriteString("Используйте стрелки ↑/↓ для прокрутки, Enter для сохранения\n")
	} else if model.State != StateError {
		// В других режимах показываем поле ввода
		s.WriteString(model.Input.View() + "\n\n")
	}

	// Отображаем сообщение об ошибке, если оно есть
	if model.ErrorMsg != "" {
		s.WriteString("\n" + FormatError(model.ErrorMsg) + "\n\n")
		s.WriteString("Нажмите Enter, чтобы повторить ввод.\n")
	}

	// Отображаем справку
	if model.ShowHelp {
		s.WriteString("\n")
		s.WriteString("Ctrl+C или Esc для выхода\n")
		s.WriteString("Ctrl+H для скрытия/показа справки\n")
		s.WriteString("Enter для подтверждения\n")
	}

	return s.String()
}
