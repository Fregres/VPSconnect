# VPSconnect

Что делает агент.

## Запуск локально

VPSCONNECT_TOKEN=test-token go run .

## API

GET /healthz
Авторизация не нужна.

GET /api/v1/status
Authorization: Bearer <token>

## Пример ответа

JSON со всеми метриками.

## Ошибки

401 — нет или неверный токен
500 — не удалось собрать метрики
