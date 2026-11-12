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

## Команды

| Команда          | Описание                                                        |
| ---------------- | --------------------------------------------------------------- |
| `debafr`         | TUI: деплой blue/green с проверками и переключением nginx       |
| `debafr status`  | Текущий деплой: версия, стратегия, порты, состояние контейнеров |
| `debafr update`  | Обновить debafr до последней версии                             |
| `debafr version` | Показать версию debafr (`--version`, `-v`)                      |

## Использование

> Обратите внимание команды выполняются для каталога: `/opt/project`

1. Структура каталога `/opt/project`
    ```bash
    compose.blue.yaml
    compose.green.yaml
    debafr.toml
    nginx.conf >> /etc/nginx/sites-available/project.ru
    ```
2. Создайте символическую ссылку на файл конфига nginx
    ```bash
    ln -sf /etc/nginx/sites-available/project.ru /opt/project/nginx.conf
    ```
3. Создайте конфигурационный файл проекта `/opt/project/debafr.toml`, пример можно посмотреть здесь: `.dev/debafr.toml`
   | Параметр | Тип | Значение по умолчанию | Обязательное |
   |----------|------|------------------------|---------------|
   | `app.project_name` | string | `"myapp"` | ✅ Да |
   | `app.images` | array of string | `["ghcr.io/app/api", "ghcr.io/app/ui", ]` | ✅ Да |
   | `app.replacer[].path` | string | `""` | ✅ Да (внутри записи) |
   | `app.replacer[].blue` | string | `""` | ✅ Да (внутри записи) |
   | `app.replacer[].green` | string | `""` | ✅ Да (внутри записи) |
   | `docker_login.enabled` | bool | `false` | Нет |
   | `docker_login.registry` | string | `""` | Нет |
   | `docker_login.username` | string | `""` | Нет |
   | `docker_login.password` | string | `""` | Нет |
   | `files.compose_blue` | string | `"compose.blue.yaml"` | Нет |
   | `files.compose_green` | string | `"compose.green.yaml"` | Нет |
   | `files.nginx_conf` | string | `"nginx.conf"` | Нет |
   | `binpaths.docker` | string | `"/usr/bin/docker"` | Нет |
   | `binpaths.curl` | string | `"/usr/bin/curl"` | Нет |
   | `binpaths.nginx` | string | `"/usr/sbin/nginx"` | Нет |
   | `nginx.enabled` | bool | `true` | Нет |
   | `nginx.test` | array of string | `["/usr/sbin/nginx", "-t"]` | Нет |
   | `nginx.reload` | array of string | `["/usr/sbin/nginx", "-s", "reload"]` | Нет |
   | `timeouts.default` | string | `"5m"` | Нет |
   | `healthcheck.max_retries` | integer | `30` | Нет |
   | `healthcheck.retry_delay` | string | `"5s"` | Нет |

    `app.replacer` — необязательный массив: на каждом update в указанном файле значение текущей стратегии заменяется на значение следующей (`blue` → `green` и наоборот). Для nginx добавьте сюда `nginx.conf`: в конфиге должна быть одна активная строка `proxy_pass` на каждый location, а записи replacer'а переключают blue- и green-адреса.

    Состав приложения задаётся не в `debafr.toml`, а labels в compose-файлах: `app.project.name` (совпадает с `app.project_name`), `app.version` и `app.service.type` — любое имя сервиса (`backend`, `frontend`, `worker`, ...). Сервисов может быть один, два или больше: healthcheck проверяет все, а при обновлении debafr останавливает все контейнеры прошлой версии. Контейнеры без `app.service.type` (например, база данных) считаются инфраструктурой и не затрагиваются.

    `nginx.test` и `nginx.reload` — массивы аргументов команды (без shell). Пустой массив пропускает команду. `nginx.enabled = false` отключает проверку версии nginx и весь шаг переключения, если nginx не используется.

4. Запустите деплой
    ```bash
    debafr
    ```
