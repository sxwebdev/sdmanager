package sdmanager

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Шаблон systemd unit
const systemdUnitTemplate = `[Unit]
Description={{.ServiceName}} Service
After=network.target

[Service]
{{ if neq .UserName "" }}User={{.UserName}}{{ end }}
WorkingDirectory={{.WorkingDirectory}}
ExecStart={{.ExecStart}}
Restart=always
RestartSec=10
OOMPolicy=restart

{{ if neq .StandardOutput "" }}StandardOutput={{.StandardOutput}}{{ end }}
{{ if neq .StandardError "" }}StandardError={{.StandardError}}{{ end }}
{{ if neq .SyslogIdentifier "" }}SyslogIdentifier={{.SyslogIdentifier}}{{ end }}

{{ if gt .MemoryMax 0 }}MemoryMax={{.MemoryMax}}M{{ end }}
{{ if gt .MemoryHigh 0 }}MemoryHigh={{.MemoryHigh}}M{{ end }}

[Install]
WantedBy=multi-user.target
`

// Получить текущую директорию
func GetCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

// Получить путь до текущего исполняемого файла
func GetCurrentExecutable() string {
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}
	return execPath
}

// Получить короткую версию пути (для placeholder)
func GetShortPath(path string) string {
	if len(path) <= 30 {
		return path
	}

	// Берем только первую и последнюю часть пути
	parts := strings.Split(path, string(os.PathSeparator))
	if len(parts) <= 2 {
		return path[:27] + "..."
	}

	return parts[0] + "/.../" + parts[len(parts)-1]
}

// Проверка валидности пути
func IsValidPath(path string) error {
	if path == "" {
		return errors.New("путь не может быть пустым")
	}

	// Проверка символов, недопустимых в пути
	for _, char := range []string{"*", "?", "<", ">", "|", ";"} {
		if strings.Contains(path, char) {
			return fmt.Errorf("путь содержит недопустимый символ: %s", char)
		}
	}

	// Проверка существования директории
	info, err := os.Stat(filepath.Dir(path))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("директория не существует: %s", filepath.Dir(path))
		}
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("путь не является директорией: %s", filepath.Dir(path))
	}

	return nil
}

// Проверка валидности имени сервиса
func IsValidServiceName(name string) error {
	if name == "" {
		return errors.New("имя сервиса не может быть пустым")
	}

	// Проверка на наличие специальных символов
	invalidChars := "*/\\|:\"<>?; "
	for _, char := range invalidChars {
		if strings.ContainsRune(name, char) {
			return fmt.Errorf("имя сервиса содержит недопустимый символ: %c", char)
		}
	}

	return nil
}

// Проверка валидности имени юзера
func IsValidUserName(name string) error {
	// Проверка на наличие специальных символов
	invalidChars := "*/\\|:\"<>?; "
	for _, char := range invalidChars {
		if strings.ContainsRune(name, char) {
			return fmt.Errorf("имя сервиса содержит недопустимый символ: %c", char)
		}
	}

	return nil
}

// Проверка и конвертация числовых значений
func ParseIntValue(value string, defaultVal int) (int, error) {
	if value == "" {
		return defaultVal, nil
	}

	val, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("введено не число")
	}

	if val < 0 {
		return 0, errors.New("значение не может быть отрицательным")
	}

	return val, nil
}

// Генерация предпросмотра unit файла
func GenerateUnitPreview(config ServiceConfig) (string, error) {
	// Подготовка шаблона
	tmpl, err := template.New("systemd-unit").Funcs(template.FuncMap{
		"gt":  func(a, b int) bool { return a > b },
		"neq": func(a, b string) bool { return a != b },
	}).Parse(systemdUnitTemplate)
	if err != nil {
		return "", fmt.Errorf("ошибка при разборе шаблона: %w", err)
	}

	// Буфер для результата
	var buf bytes.Buffer

	// Приведение имени сервиса к формату с заглавной буквы
	caser := cases.Title(language.English)

	// Формирование данных для шаблона
	data := struct {
		ServiceName      string
		UserName         string
		WorkingDirectory string
		ExecStart        string
		StandardOutput   string
		StandardError    string
		SyslogIdentifier string
		MemoryHigh       int
		MemoryMax        int
	}{
		ServiceName:      caser.String(config.ServiceName),
		UserName:         config.UserName,
		WorkingDirectory: config.WorkingDirectory,
		ExecStart:        config.ExecStart,
		StandardOutput:   config.StandardOutput,
		StandardError:    config.StandardError,
		SyslogIdentifier: config.SyslogIdentifier,
		MemoryHigh:       config.MemoryHigh,
		MemoryMax:        config.MemoryMax,
	}

	// Выполнение шаблона
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("ошибка при выполнении шаблона: %w", err)
	}

	serviceUnit := removeMultipleLines(buf.String())

	return serviceUnit, nil
}

// Создание unit-файла
func CreateUnitFile(config ServiceConfig, overwrite bool) error {
	unitFilePath := filepath.Join(config.UnitFilePath, config.ServiceName+".service")

	// Проверяем, существует ли файл и нужно ли его перезаписывать
	if !overwrite {
		if _, err := os.Stat(unitFilePath); err == nil {
			return fmt.Errorf("файл %s уже существует и не будет перезаписан", unitFilePath)
		}
	}

	// Получаем предпросмотр содержимого
	content, err := GenerateUnitPreview(config)
	if err != nil {
		return err
	}

	// Создание файла
	file, err := os.Create(unitFilePath)
	if err != nil {
		return fmt.Errorf("ошибка при создании файла: %w", err)
	}
	defer file.Close()

	// Запись содержимого
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("ошибка при записи в файл: %w", err)
	}

	return nil
}

// Выполнение системных команд с выводом результата
func ExecuteCommand(name string, args ...string) (string, error) {
	// Проверяем наличие исполняемого файла
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("команда %s не найдена. Убедитесь, что системный сервис установлен и путь корректен", name)
	}

	cmd := exec.Command(path, args...)
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		if outputStr == "" {
			return outputStr, fmt.Errorf("ошибка при выполнении команды %s: %w", name, err)
		}
		return outputStr, fmt.Errorf("%w: %s", err, outputStr)
	}

	return outputStr, nil
}

// Выполнение команды daemon-reload
func ReloadDaemon() (string, error) {
	return ExecuteCommand("systemctl", "daemon-reload")
}

// Выполнение команды enable
func EnableService(serviceName string) (string, error) {
	return ExecuteCommand("systemctl", "enable", serviceName)
}

// Выполнение команды start
func StartService(serviceName string) (string, error) {
	output, err := ExecuteCommand("systemctl", "start", serviceName)
	if err != nil {
		return "", err
	}

	if output != "" {
		return fmt.Sprintf("Сервис %s успешно запущен\n%s", serviceName, output), nil
	}
	return fmt.Sprintf("Сервис %s успешно запущен", serviceName), nil
}

// Выполнение команды stop
func StopService(serviceName string) (string, error) {
	output, err := ExecuteCommand("systemctl", "stop", serviceName)
	if err != nil {
		return "", err
	}

	if output != "" {
		return fmt.Sprintf("Сервис %s успешно остановлен\n%s", serviceName, output), nil
	}
	return fmt.Sprintf("Сервис %s успешно остановлен", serviceName), nil
}

// Выполнение команды restart
func RestartService(serviceName string) (string, error) {
	output, err := ExecuteCommand("systemctl", "restart", serviceName)
	if err != nil {
		return "", err
	}

	if output != "" {
		return fmt.Sprintf("Сервис %s успешно перезапущен\n%s", serviceName, output), nil
	}
	return fmt.Sprintf("Сервис %s успешно перезапущен", serviceName), nil
}

// Выполнение просмотра логов
func ViewServiceLogs(ctx context.Context, serviceName string) error {
	cmd := exec.CommandContext(ctx, "journalctl", "-n", "50", "-u", serviceName, "--output=json", "--no-pager")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ошибка запуска: %w", err)
	}

	done := make(chan struct{})
	go func() {
		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {
			jsonLine := scanner.Text()

			var entry map[string]interface{}
			if err := json.Unmarshal([]byte(jsonLine), &entry); err != nil {
				fmt.Fprintf(os.Stderr, "ошибка парсинга JSON: %v\n", err)
				continue
			}

			if message, ok := entry["MESSAGE"].(string); ok {
				fmt.Println(message)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "ошибка чтения вывода: %v\n", err)
		}

		close(done)
	}()

	select {
	case <-done:
		return cmd.Wait()
	case <-ctx.Done():
		return cmd.Process.Kill()
	}
}

// Полностью установить сервис (создать файл, reload, enable, start)
func InstallService(config ServiceConfig, actions UserActions) (string, error) {
	var resultMessages []string

	// 1. Создаем unit-файл
	unitFilePath := filepath.Join(config.UnitFilePath, config.ServiceName+".service")
	err := CreateUnitFile(config, actions.Overwrite)
	if err != nil {
		return "", fmt.Errorf("ошибка при создании unit-файла: %w", err)
	}

	// Получаем абсолютный путь
	absPath, err := filepath.Abs(unitFilePath)
	if err != nil {
		absPath = unitFilePath
	}
	resultMessages = append(resultMessages, "\n\nSystemd unit файл создан: "+absPath)

	// 2. Если выбрано, выполняем daemon-reload
	if actions.ReloadDaemon {
		output, err := ReloadDaemon()
		if err != nil {
			return strings.Join(resultMessages, "\n"), err
		}
		if output != "" {
			resultMessages = append(resultMessages, output)
		}
		resultMessages = append(resultMessages, "Systemd daemon перезагружен")
	}

	// 3. Если выбрано, выполняем enable
	if actions.EnableService {
		output, err := EnableService(config.ServiceName)
		if err != nil {
			return strings.Join(resultMessages, "\n"), err
		}
		if output != "" {
			resultMessages = append(resultMessages, output)
		}
		resultMessages = append(resultMessages, "Сервис активирован (enabled)")
	}

	// 4. Если выбрано, выполняем start
	if actions.StartService {
		output, err := ExecuteCommand("systemctl", "start", config.ServiceName)
		if err != nil {
			return strings.Join(resultMessages, "\n"), err
		}
		if output != "" {
			resultMessages = append(resultMessages, output)
		}
		resultMessages = append(resultMessages, "Сервис запущен (started)")
	}

	resultMessages = append(resultMessages, "Установка успешно завершена")
	return strings.Join(resultMessages, "\n"), nil
}

// Проверить существует ли файл
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// removeMultipleLines принимает исходный текст и возвращает текст, в котором
// между непустыми строками не более одной пустой строки.
func removeMultipleLines(input string) string {
	lines := strings.Split(input, "\n")
	var result []string
	prevBlank := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			// Если предыдущая строка (из исходного массива) начинается с "[", пропускаем пустую строку.
			if i > 0 && strings.HasPrefix(strings.TrimSpace(lines[i-1]), "[") {
				continue
			}
			// Добавляем пустую строку, если предыдущей не было.
			if !prevBlank {
				result = append(result, "")
				prevBlank = true
			}
		} else {
			result = append(result, line)
			prevBlank = false
		}
	}

	return strings.Join(result, "\n")
}
