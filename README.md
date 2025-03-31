# Systemd Service Manager

## 🚀 Project Description

Systemd Service Manager is an interactive command-line utility for managing systemd services in Linux, designed to simplify system service administration. The tool provides an intuitive interface for performing essential service operations.

## ✨ Features

### 🔧 **Service Management**

- Start services
- Stop services
- Restart services
- View service logs

### 📦 **New Service Installation**

- Interactive systemd unit file creation
- Flexible service parameter configuration
- Support for advanced configuration options

### 🛡️ **Advanced Configuration Capabilities**

- Set working directory
- Configure start command
- Memory usage limitations (MemoryHigh and MemoryMax)
- Customize unit file path

### 🖥️ **User-Friendly Interface**

- Interactive menu
- Step-by-step service configuration
- Unit file preview before creation

## 🛠️ Requirements

- Linux with systemd
- Go 1.24+
- Superuser privileges (sudo) for system service management

## 🚀 Installation

### From Source Code

```bash
git clone https://github.com/sxwebdev/sdmanager.git
cd systemd
go build -o bin/sdmanager ./cmd/sdmanager
sudo ./bin/sdmanager
```

### Or Install via go install

```bash
go install github.com/sxwebdev/sdmanager/cmd/sdmanager@latest
```

### Or Install via script

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/sxwebdev/sdmanager/refs/heads/master/scripts/install.sh)"
```

## 📖 User Guide

### Running the Utility

```bash
sudo ./sdmanager
```

### Main Functions

1. **Start a Service**

   - Select "Start Service"
   - Enter the service name

2. **Stop a Service**

   - Select "Stop Service"
   - Enter the service name

3. **Install a New Service**

   - Select "Install Service"
   - Follow interactive prompts:
     - Enter service name
     - Specify working directory
     - Configure start command
     - Set memory limitations (optional)
     - Set CPU usage limit in percents (optional)
     - Set allowed CPU Cores to use in system (optional)
     - Choose additional options

4. **View Logs**
   - Select "View Logs"
   - Enter the service name

## 🌟 Advantages

- **Ease of Use**: Intuitive command-line interface
- **Safety**: Preview and confirm actions before execution
- **Flexibility**: Comprehensive service configuration options
- **Performance**: Quick service operation execution

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📜 License

Distributed under the MIT License. See `LICENSE` for more information.

## 🐛 Bug Reports

Please report bugs through the GitHub Issues section.

---

**Note**: This utility requires caution when working with system services. Always verify the consequences before performing actions.
