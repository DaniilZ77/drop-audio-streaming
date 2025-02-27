# Drop Beat Service [![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Микросервис для стриминга аудио, загрузки, обновления, приобретения битов

## Основные возможности

- Создание бита (с жанрами, тэгами, настроениями и другими многочисленными настройками через администратора)
- Множественная фильтрация битов по различным параметрам
- Удаление/изменения бита (через администратора)
- Приобретение бита (через администратора)
- Стриминг аудио контента

## Стек

- Go
- PostgreSQL (pgx + sqlc)
- Minio (для хранения битов, архивов, фотографий)
- Docker
- gRPC + gRPC Gateway

## Запуск сервиса

1. Клонируйте репозиторий:
   ```bash
   $ git clone https://github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming.git

   $ cd drop-audio-streaming
   ```

2. Запустите сервис:
   ```bash
   $ docker compose up
   ```

## База данных

Схема базы данных находится на следующем ресурсе:

- https://dbdiagram.io/d/drop-beats-67c0c75b263d6cf9a0b96ccd

## Прото-файлы

https://github.com/MAXXXIMUS-tropical-milkshake/drop-protos

Для более подробного ознакомления с API следует читать документацию (в `docs`).
