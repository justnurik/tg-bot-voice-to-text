# Telegram Voice-to-Text Bot (@voicetotextnurik\_bot)

This project is a Telegram bot that transcribes voice messages into text. The backend is written in Go, and the speech recognition models run as Python microservices using [whisper-instance-manager](https://github.com/justnurik/whisper-instance-manager).

## Repositories

- Bot (this repository): [`tg-bot-voice-to-text`](https://github.com/justnurik/tg-bot-voice-to-text)
- Whisper instances: [`whisper-instance-manager`](https://github.com/justnurik/whisper-instance-manager)

## Features

- Transcription of voice messages via Whisper
- Webhook-based interaction with Telegram API
- Configuration via YAML
- Easy build and launch
- Supports multiple model instances
- Logging and future support for product/technical/ML metrics

## Project Structure

```
.
├── bin/                    # Compiled Go binary
├── config.yml              # Configuration file
├── run.py                 # Python launcher script
├── src/                   # Go source code
├── webhook.pem  .key     # TLS certificate and key
├── openssl.cnf            # OpenSSL config
└── logs/, downloads/, ... # Other folders
```

## Configuration

All settings are managed in `config.yml`:

```yaml
api_token: "YOUR_TELEGRAM_BOT_TOKEN"
host_url: "https://YOUR_PUBLIC_IP"
listen_port: 443
cache_size: 10000
log_file: "logs/bot.log"
log_level: "info"
debug: false
model_instance_urls:
  - "http://localhost:9000/inference"
```

## Certificates

Telegram webhooks require HTTPS. A self-signed certificate is used:

- `webhook.pem` — public certificate
- `webhook.key` — private key

> Your server's IP must be included in the **Subject Alternative Name (SAN)** field.

Generate a certificate like this:

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout webhook.key \
  -out webhook.pem \
  -config openssl.cnf \
  -extensions req_ext
```

Example `openssl.cnf`:

```ini
[req]
default_bits       = 2048
prompt             = no
default_md         = sha256
req_extensions     = req_ext
distinguished_name = dn

[dn]
CN = YOUR_PUBLIC_IP

[req_ext]
subjectAltName = @alt_names

[alt_names]
IP.1 = YOUR_PUBLIC_IP
```

## 🚀 Run

1. Place `webhook.pem` and `webhook.key` in the project root
2. Edit `config.yml`
3. Launch with:

```bash
python3 run.py
```

## Metrics (in progress)

Planned support for:

### Technical

- Model response time
- Request volume
- Errors
- Load (Prometheus)

### ML

- Latency
- Error rate
- Throughput

### Product

- MAU, DAU, Retention
- Active users
- Engagement depth

## Telegram Bot

Production bot: [@voicetotextnurik\_bot](https://t.me/voicetotextnurik_bot)

## Requirements

- Go 1.20+
- Python 3.8+
- Whisper instance manager

## Feedback

PRs and ideas welcome. Use GitHub Issues for bugs and feature requests.

# Telegram Voice-to-Text Bot (@voicetotextnurik_bot)

Этот проект представляет собой Telegram-бот, который преобразует голосовые сообщения в текст. Backend написан на Go, а модели распознавания речи работают как микросервисы на Python с использованием [whisper-instance-manager](https://github.com/justnurik/whisper-instance-manager).

## Репозитории

- Бот (этот репозиторий): [`tg-bot-voice-to-text`](https://github.com/justnurik/tg-bot-voice-to-text)
- Экземпляры Whisper: [`whisper-instance-manager`](https://github.com/justnurik/whisper-instance-manager)

## Возможности

- Транскрипция голосовых сообщений с помощью Whisper
- Взаимодействие с Telegram API через вебхуки
- Конфигурация через YAML
- Простая сборка и запуск
- Поддержка нескольких экземпляров моделей
- Логирование и будущая поддержка продуктовых/технических/ML-метрик

## Структура проекта

```
.
├── bin/                    # Скомпилированный бинарный файл Go
├── config.yml              # Файл конфигурации
├── run.py                 # Скрипт запуска на Python
├── src/                   # Исходный код на Go
├── webhook.pem  .key     # TLS-сертификат и ключ
├── openssl.cnf            # Конфигурация OpenSSL
└── logs/, downloads/, ... # Другие папки
```

## Конфигурация

Все настройки задаются в файле `config.yml`:

```yaml
api_token: "YOUR_TELEGRAM_BOT_TOKEN"
host_url: "https://YOUR_PUBLIC_IP"
listen_port: 443
cache_size: 10000
log_file: "logs/bot.log"
log_level: "info"
debug: false
model_instance_urls:
  - "http://localhost:9000/inference"
```

## Сертификаты

Вебхуки Telegram требуют HTTPS. Используется самоподписанный сертификат:

- `webhook.pem` — публичный сертификат
- `webhook.key` — приватный ключ

> IP-адрес вашего сервера должен быть указан в поле **Subject Alternative Name (SAN)**.

Генерация сертификата:

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout webhook.key \
  -out webhook.pem \
  -config openssl.cnf \
  -extensions req_ext
```

Пример `openssl.cnf`:

```ini
[req]
default_bits       = 2048
prompt             = no
default_md         = sha256
req_extensions     = req_ext
distinguished_name = dn

[dn]
CN = YOUR_PUBLIC_IP

[req_ext]
subjectAltName = @alt_names

[alt_names]
IP.1 = YOUR_PUBLIC_IP
```

## 🚀 Запуск

1. Поместите `webhook.pem` и `webhook.key` в корень проекта
2. Отредактируйте `config.yml`
3. Запустите с помощью:

```bash
python3 run.py
```

## Метрики (в разработке)

Планируется поддержка следующих метрик:

### Технические

- Время ответа модели
- Объем запросов
- Ошибки
- Нагрузка (Prometheus)

### ML

- Задержка
- Уровень ошибок
- Пропускная способность

### Продуктовые

- MAU, DAU, Retention
- Активные пользователи
- Глубина вовлеченности

## Telegram-бот

Продакшн-бот: [@voicetotextnurik_bot](https://t.me/voicetotextnurik_bot)

## Требования

- Go 1.20+
- Python 3.8+
- Whisper instance manager

## Обратная связь

Приветствуются пул-реквесты и идеи. Используйте GitHub Issues для сообщений об ошибках и запросов новых функций.