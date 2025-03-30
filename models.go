package sdmanager

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

// Режимы приложения
const (
	ModeMainMenu = iota
	ModeInstallService
	ModeServiceInput
	ModeExit  // Режим выхода из приложения
	ModeError // Режим отображения ошибки
)

// Действия с сервисами
const (
	ActionStart   = "start"
	ActionStop    = "stop"
	ActionRestart = "restart"
	ActionViewLog = "log"
)

// Пункты меню
type MenuAction string

const (
	ActionStartService   MenuAction = "Запустить сервис"
	ActionStopService    MenuAction = "Остановить сервис"
	ActionRestartService MenuAction = "Перезапустить сервис"
	ActionViewLogs       MenuAction = "Просмотр логов"
	ActionInstallService MenuAction = "Установить сервис"
	ActionExit           MenuAction = "Выход"
)

// Состояния установки сервиса
const (
	StateServiceName = iota
	StateUserName
	StateWorkingDirectory
	StateExecStart
	StateStandardOutput
	StateStandardError
	StateSyslogIdentifier
	StateMemoryHigh
	StateMemoryMax
	StateUnitLocation
	StateOverwrite
	StateOptionsSelect
	StatePreviewUnit
	StateDone
	StateError
)

// Пункт меню
type MenuItem struct {
	Title  string
	Action MenuAction
}

func (i MenuItem) FilterValue() string { return i.Title }

// Опция для выбора
type Option struct {
	Name     string
	Selected bool
}

// Данные для создания и настройки systemd unit
type ServiceConfig struct {
	ServiceName      string
	UserName         string
	WorkingDirectory string
	ExecStart        string
	StandardOutput   string
	StandardError    string
	SyslogIdentifier string
	MemoryHigh       int
	MemoryMax        int
	UnitFilePath     string
}

// Действия пользователя
type UserActions struct {
	Overwrite     bool
	ReloadDaemon  bool
	EnableService bool
	StartService  bool
}

// Модель меню
type MenuModel struct {
	List     list.Model
	Choice   MenuAction
	Quitting bool
}

// Модель ввода имени сервиса
type ServiceInputModel struct {
	Input     textinput.Model
	Action    string
	Message   string
	Error     string
	ResultMsg string
	Quitting  bool
}

// Модель для установки сервиса
type InstallModel struct {
	State          int
	Config         ServiceConfig
	Actions        UserActions
	Input          textinput.Model
	Viewport       viewport.Model
	Message        string
	ErrorMsg       string
	PreviewContent string
	ShowHelp       bool
	Quitting       bool
	Aborted        bool
	ResultMsg      string
	Options        []Option
	CurrentOption  int
}

// Основная модель приложения
type AppModel struct {
	Mode              int
	MenuModel         MenuModel
	InstallModel      InstallModel
	ServiceInputModel ServiceInputModel
	Message           string
	Error             string
	FatalError        bool

	options AppOptions
}
