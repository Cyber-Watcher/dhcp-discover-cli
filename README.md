# DHCP Discover CLI

A lightweight cross-platform command-line utility to detect DHCP servers in your network. It sends DHCPDISCOVER packets and listens for DHCPOFFER responses to identify active DHCP servers. Ideal for network administrators and security professionals to quickly identify rogue DHCP servers.

## Features

- **Cross-platform**: Runs on Windows, Linux, and macOS
- **Simple interface**: Easy-to-use command-line options
- **Multiple discovery attempts**: Configurable retry mechanism
- **Verbose logging**: Detailed output for troubleshooting
- **No dependencies**: Single static binary

## Installation

Download pre-built binaries from the [Releases](https://github.com/Cyber-Watcher/dhcp-discover-cli/releases) page and unzip it.

## Usage

### Basic detection
```bash
sudo dhcp-discover
```

### Show available interfaces
```bash
dhcp-discover --show-interfaces
```

### Select specific interface
```bash
sudo dhcp-discover --iface eth0
# OR
sudo dhcp-discover --iface-index 2
```

### Advanced options
```bash
sudo dhcp-discover \
  --timeout 10s \
  --retry 5 \
  --verbose
```

### Help
```bash
dhcp-discover --help
```

## Options
| Option               | Description                                 |
|----------------------|---------------------------------------------|
| `--iface`            | Select interface by name (e.g., eth0)       |
| `--iface-index`      | Select interface by number from list        |
| `--show-interfaces`  | List available network interfaces           |
| `--timeout`          | Response timeout (default: 8s)              |
| `--retry`            | Number of discovery attempts (default: 3)   |
| `--verbose`          | Enable detailed logging to console and file |

## Building from Source

1. Clone repository:
```bash
git clone https://github.com/Cyber-Watcher/dhcp-discover-cli.git
cd dhcp-discover-cli
```

2. Build:
```bash
go build -o dhcp-discover ./cmd/dhcp-discover-cli
```

3. Run:
```bash
sudo ./dhcp-discover
```

## License
MIT License - see [LICENSE](LICENSE) for details

---

# DHCP Discover CLI (Russian)

Легковесная кроссплатформенная утилита командной строки для обнаружения DHCP-серверов в сети. Отправляет DHCPDISCOVER пакеты и прослушивает DHCPOFFER ответы для выявления активных DHCP-серверов. Идеально подходит для системных администраторов и специалистов по безопасности для быстрого выявления неавторизованных DHCP-серверов.

## Особенности

- **Кроссплатформенность**: Работает на Windows, Linux и macOS
- **Простой интерфейс**: Легкие в использовании параметры командной строки
- **Несколько попыток обнаружения**: Настраиваемый механизм повторов
- **Подробное логирование**: Детальный вывод для диагностики
- **Без зависимостей**: Один статический бинарный файл

## Установка

Скачайте готовые бинарные файлы из раздела [Releases](https://github.com/Cyber-Watcher/dhcp-discover-cli/releases) и разархивируйте.

## Использование

### Базовое обнаружение
```bash
sudo dhcp-discover
```

### Показать доступные интерфейсы
```bash
dhcp-discover --show-interfaces
```

### Выбор конкретного интерфейса
```bash
sudo dhcp-discover --iface eth0
# ИЛИ
sudo dhcp-discover --iface-index 2
```

### Дополнительные опции
```bash
sudo dhcp-discover \
  --timeout 10s \
  --retry 5 \
  --verbose
```

### Справка
```bash
dhcp-discover --help
```

## Параметры
| Параметр              | Описание                                      |
|-----------------------|-----------------------------------------------|
| `--iface`             | Выбрать интерфейс по имени (например, eth0)  |
| `--iface-index`       | Выбрать интерфейс по номеру из списка        |
| `--show-interfaces`   | Показать доступные сетевые интерфейсы        |
| `--timeout`           | Таймаут ожидания ответа (по умолчанию: 8с)   |
| `--retry`             | Количество попыток (по умолчанию: 3)         |
| `--verbose`           | Подробный вывод в консоль и файл             |

## Сборка из исходников

1. Клонируйте репозиторий:
```bash
git clone https://github.com/Cyber-Watcher/dhcp-discover-cli.git
cd dhcp-discover-cli
```

2. Соберите проект:
```bash
go build -o dhcp-discover ./cmd/dhcp-discover-cli
```

3. Запустите:
```bash
sudo ./dhcp-discover
```

## Лицензия
MIT License - подробнее в файле [LICENSE](LICENSE)