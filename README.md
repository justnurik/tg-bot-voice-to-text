# Telegram Voice-to-Text Bot

## Описание

Этот проект — Telegram-бот, который преобразует голосовые сообщения в текст с использованием моделей распознавания речи, таких как OpenAI Whisper или другие. Бот контейнеризирован с помощью Docker, настраивается через конфигурационные файлы и поддерживает сбор продуктовых и технических метрик. Масштабируемость реализована через добавление инстансов моделей с автоматическим распределением нагрузки.

## Основные функции

- **Транскрибация голоса**: Преобразование голосовых сообщений в текст.
- **Гибкая конфигурация**: Настройки задаются в конфигурационных файлах.
- **Docker**: Удобное развёртывание в контейнерах.
- **Метрики**: Сбор данных о производительности и использовании.
- **Масштабируемость**: Автоматическое распределение нагрузки между инстансами моделей.

## Требования

- Docker и Docker Compose
- API-ключ Telegram
- Модели распознавания речи (например, OpenAI API или локальные файлы)

## Установка и запуск

1. **Клонируйте репозиторий**:
   ```bash
   git clone https://github.com/your-repo/telegram-voice-to-text-bot.git
   cd telegram-voice-to-text-bot
   ```

2. **Настройте конфигурацию**:
   - `config/bot.yml`: Токен Telegram, логирование и др.
   - `config/models.yml`: Список моделей и их параметры.

3. **Запустите контейнер**:
   ```bash
   docker-compose up --build
   ```

## Конфигурация моделей

Бот поддерживает локальные модели и API. Примеры:

- **OpenAI Whisper (API)**:
  ```yaml
  models:
    - name: openai_whisper
      type: api
      api_key: your_openai_api_key
      endpoint: https://api.openai.com/v1/audio/transcriptions
  ```

- **Локальная модель (Whisper)**:
  ```yaml
  models:
    - name: whisper_local
      type: local
      path: /models/whisper/model.bin
  ```

### Источники моделей
- **OpenAI Whisper**: Используйте API или скачайте локальную модель с [GitHub OpenAI Whisper](https://github.com/openai/whisper).
- **Другие модели**: Например, [Mozilla DeepSpeech](https://github.com/mozilla/DeepSpeech) — скачайте с официального репозитория.

## Метрики

- **Продуктовые**: Количество сообщений, время обработки, ошибки.
- **Технические**: Загрузка CPU, памяти, отклик моделей.  
Доступны через `/metrics` (интеграция с Prometheus/Grafana).

## Масштабирование

1. Добавьте инстанс в `config/models.yml`.
2. Перезапустите:
   ```bash
   docker-compose down && docker-compose up -d
   ```

Бот сам распределит нагрузку.

## Лицензия

MIT License. См. [LICENSE](./LICENSE).
