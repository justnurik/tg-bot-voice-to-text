# Telegram Voice-to-Text Bot (@voicetotextnurik_bot)

Этот проект представляет собой Telegram-бот, который преобразует голосовые сообщения в текст. Backend написан на Go, а модели распознавания речи работают как микросервисы на Python с использованием [whisper-instance-manager](https://github.com/justnurik/whisper-instance-manager).

## Репозитории

- Бот (этот репозиторий): [`tg-bot-voice-to-text`](https://github.com/justnurik/tg-bot-voice-to-text)
- Экземпляры Whisper: [`whisper-instance-manager`](https://github.com/justnurik/whisper-instance-manager)

## Возможности

- Транскрипция голосовых сообщений с помощью Whisper
- Взаимодействие с Telegram API через вебхуки
- Кэширование результатов обработки
- Конфигурация через YAML
- Простая сборка и запуск
- Поддержка нескольких экземпляров моделей
- Логгирование через Uber/zap
- Будущая поддержка продуктовых/технических/ML-метрик

## Структура проекта

```
.
├── bin/                    # Скомпилированные бинарные файлы
├── cmd/
│   └── vtt/                # Основное приложение
├── configs/                # Конфигурационные файлы
│   ├── bot.yml             # Конфигурация бота
│   └── logger.yml          # Конфигурация логгера
├── internal/               # Внутренние пакеты
│   └── vtt/                # Логика бота
├── pkg/                    # Вспомогательные пакеты
│   ├── botwork/            # Работа с Telegram API
│   ├── cache/              # Кэширование
│   ├── queue/              # Очереди
│   ├── scheduler/          # Планировщики задач
│   ├── setup/              # Утилиты инициализации
│   └── utils/              # Вспомогательные утилиты
├── Makefile                # Скрипты сборки
├── go.mod                  # Зависимости Go
├── go.sum                  # Контрольные суммы зависимостей
└── README.md
```

## Конфигурация

Конфигурация разделена на два файла:

configs/bot.yml
```yaml
token: "YOUR_TELEGRAM_BOT_TOKEN"
mode: "webhook" # или "longpoll"
name: "VoiceToTextBot"
debug: false
listen_addr: ":8080"
cache_size: 1000
timeout: 60
model_instance_urls:
  - "http://localhost:9000/transcriptions"
  - "http://another-instance:9000/transcriptions"
```

configs/logger.yml
```yaml
add-stacktrace: true
stacktrace-log-level: error
add-caller: true
console: true
console-log-level: info
log-files-config:
- file-path: logs/bot.log
  log-level: info
  max-size: 1024
  max-backups: 2
  max-age: 30
```

## Запуск программы бота

Установите зависимости:

```bash
go mod tidy
```

Соберите бинарник:
```bash
make build
```
Запустите бота:
```bash
./bin/vtt # можно явно указать откуда брать конфиги: --bot-config=configs/bot.yml --logger-config=configs/logger.yml
```

## Запуск бота

### Обеспечение HTTPS доступа
Для работы с Telegram API через вебхуки необходимо обеспечить HTTPS соединение. Вот основные способы:

1. Использование ngrok (для разработки/тестирования)
```bash
# Установите ngrok: https://ngrok.com/download
ngrok http 8080
```
После запуска вы получите HTTPS-ссылку вида https://<random-id>.ngrok-free.app

2. Настройка Nginx с Let's Encrypt (продакшн)
```bash
# Установите Nginx и Certbot
sudo apt install nginx certbot python3-certbot-nginx

# Создайте конфиг для домена
sudo nano /etc/nginx/sites-available/yourdomen.com

# Получите сертификат
sudo certbot --nginx -d yourdomen.com

# Перезапустите Nginx
sudo systemctl reload nginx
```

3. Использование Docker с Traefik
Пример docker-compose.yml:
```yaml
version: '3'
services:
  bot:
    image: your-bot-image
    ports:
      - "8080:8080"
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.bot.rule=Host(`yourdomain.com`)"
      - "traefik.http.routers.bot.entrypoints=websecure"
      - "traefik.http.routers.bot.tls.certresolver=myresolver"
```

### Установка вебхука

После настройки HTTPS зарегистрируйте вебхук в Telegram API:
```bash
curl -F "url=https://yourdomen.com/voice-to-text-bot-webhook/webhook" \
  "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/setWebhook"
```

Проверьте статус вебхука:
```bash
curl "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getWebhookInfo"
```


## Запуск в продакшн
Рекомендуемые способы запуска:

  1. Systemd сервис
Создайте файл /etc/systemd/system/vtt-bot.service:
```ini
[Unit]
Description=Voice-to-Text Telegram Bot
After=network.target

[Service]
User=tgbotvtt
Group=tgbotvtt
WorkingDirectory=/home/youruser/tg-bot-voice-to-text
ExecStart=/home/youruser/tg-bot-voice-to-text/bin/vtt
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

2. Команды управления:
```bash
sudo systemctl daemon-reload
sudo systemctl enable vtt-bot
sudo systemctl start vtt-bot
sudo journalctl -u vtt-bot -f
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

## Telegram-бот

Продакшн-бот: [@voicetotextnurik_bot](https://t.me/voicetotextnurik_bot)

## Требования

- Go 1.20+
- Python 3.8+
- Whisper instance manager

## Лицензия

Проект распространяется под лицензией [MIT](https://github.com/justnurik/tg-bot-voice-to-text/blob/main/LICENSE)
