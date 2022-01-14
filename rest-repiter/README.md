# Autotest Tools

В этом репозитарии некоторые из модулей которые вошли в состав программного обеспечения Сивилла предназначенного для 

Получить
```sh
curl -H "Content-Type: application/json" -X POST http://localhost:7456/sh -d "Hello World"
```
`
{
	"service": "go.tracker.svc.capproc",
    "endpoint": "Capproc.GetCapsStates",
    "request": {
		"offerId": 90
	}
}
`

CGO_ENABLED=0 GOOS=linux go build -gcflags "all=-N -l" -o service