[![audit](https://github.com/dimns/debafr/actions/workflows/audit.yml/badge.svg?branch=master)](https://github.com/dimns/debafr/actions/workflows/audit.yml)

# Debafr

Deploy Backend/Frontend application

## Установка

```bash
# Установка последней версии
curl -LsSf https://raw.githubusercontent.com/dimns/debafr/refs/heads/master/scripts/install.sh | bash

# Установка конкретной версии
curl -LsSf https://raw.githubusercontent.com/dimns/debafr/refs/heads/master/scripts/install.sh | bash -s -- -v v0.6.0

# Тестовый прогон (без изменений)
curl -LsSf https://raw.githubusercontent.com/dimns/debafr/refs/heads/master/scripts/install.sh | bash -s -- --dry-run

# Удаление
curl -LsSf https://raw.githubusercontent.com/dimns/debafr/refs/heads/master/scripts/install.sh | bash -s -- --uninstall

# Тихая установка
curl -LsSf https://raw.githubusercontent.com/dimns/debafr/refs/heads/master/scripts/install.sh | bash -s -- --quiet
```

## Использование

> Обратите внимание команды выполняются для каталога: `/opt/project`

Структура каталога `/opt/project`

```
compose.blue.yaml
compose.green.yaml
debafr.toml
nginx.conf >> /etc/nginx/sites-available/project.ru
```

1. Создайте символическую ссылку на файл конфига nginx
    ```bash
    ln -sf /etc/nginx/sites-available/project.ru /opt/project/nginx.conf
    ```
2. Создайте конфигурационный файл проекта `/opt/project/debafr.toml`, пример можно посмотреть здесь: `.dev/debafr.toml`
3. Запустите приложение
    ```bash
    debafr
    ```
