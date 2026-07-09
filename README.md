# VPSconnect 
## Запуск локально

VPSCONNECT_TOKEN=test-token go run .

## API

GET /healthz
Авторизация не нужна.

GET /api/v1/status
`Authorization: Bearer <token>`

## Пример ответа

curl --noproxy '*' -i \               
  -H "Authorization: Bearer test-token" \
  http://127.0.0.1:6767/api/v1/status

```
HTTP/1.1 200 OK
Cache-Control: no-store
Content-Type: application/json
Date: Thu, 09 Jul 2026 XX:XX:XX GMT
Content-Length: 302

{
"memory":
	{"total_bytes":14554492928,"available_bytes":10267619328,"used_bytes":4286873600},
"uptime":
	{"seconds":15583},
"collected_at":	
	"2026-07-09T12:37:01.523727671Z",
"disk":	
	{"total_bytes":246574563328,"available_bytes":71733608448,"used_bytes":162241040384},
"cpu":	
	{"usage_percent":13.77551020408163}
}

```

## Ошибки

401 — нет или неверный токен
500 — не удалось собрать метрики
