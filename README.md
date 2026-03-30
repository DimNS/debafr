[![audit](https://github.com/dimns/debafr/actions/workflows/audit.yml/badge.svg?branch=master)](https://github.com/dimns/debafr/actions/workflows/audit.yml)

# Debafr

Deploy Backend/Frontend application

## Установка

```bash
# Установка последней версии
curl -LsSf https://raw.githubusercontent.com/dimns/debafr/refs/heads/master/scripts/install.sh | sh

# Установка конкретной версии
curl -LsSf https://raw.githubusercontent.com/dimns/debafr/refs/heads/master/scripts/install.sh | sh -s -- -v v0.6.0

# Тестовый прогон (без изменений)
curl -LsSf https://raw.githubusercontent.com/dimns/debafr/refs/heads/master/scripts/install.sh | sh -s -- --dry-run

# Удаление
curl -LsSf https://raw.githubusercontent.com/dimns/debafr/refs/heads/master/scripts/install.sh | sh -s -- --uninstall

# Тихая установка
curl -LsSf https://raw.githubusercontent.com/dimns/debafr/refs/heads/master/scripts/install.sh | sh -s -- --quiet
```

## Использование

> Обратите внимание команды выполняются для каталога: `/opt/project`

1. Создайте символическую ссылку на файл конфига nginx
    ```bash
    ln -sf /etc/nginx/sites-available/project.ru /opt/project/nginx.conf
    ```
2. Создайте конфигурационный файл проекта, пример можно посмотреть здесь: `.dev/.debafr.toml`
3. Запустите приложение
    ```bash
    debafr
    ```
