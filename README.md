# go-http

![tests action](https://github.com/apoldev/go-http/actions/workflows/tests.yml/badge.svg?branch=master)


Задача описана [тут](#hh)

___ 


### Как запустить

```bash
docker-compose up
```

### Запуск тестов 

```bash
go test -cover  ./...
```

### Реализация Limiter
___

Ограничить общее количество одновременных запросов можно с помощью `semaphore`
Для реалиации выбраны каналы и атомики.

Тут небольшой бенчмарк, для сравнения скорости работы каналов и атомиков.

```bash
BenchmarkChanAtom/chan_1000-8            2940975               403.1 ns/op            64 B/op          1 allocs/op
BenchmarkChanAtom/atom_1000-8            3278354               368.2 ns/op            64 B/op          1 allocs/op
```


<h3 id="hh">Задача HTTP-мультиплексор</h3>

___ 

- приложение представляет собой http-сервер с одним хендлером
- хендлер на вход получает POST-запрос со списком url в json-формате
- сервер запрашивает данные по всем этим url и возвращает результат клиенту в json-формате
- если в процессе обработки хотя бы одного из url получена ошибка, обработка всего списка прекращается и клиенту возвращается текстовая ошибка

### Ограничения:
- использовать можно только компоненты стандартной библиотеки Go
- сервер не принимает запрос если количество url в в нем больше 20
- сервер не обслуживает больше чем 100 одновременных входящих http-запросов
- для каждого входящего запроса должно быть не больше 4 одновременных исходящих
- таймаут на запрос одного url - секунда
- обработка запроса может быть отменена клиентом в любой момент, это должно повлечь за собой остановку всех операций связанных с этим запросом
- сервис должен поддерживать 'graceful shutdown'
- результат должен быть выложен на github

### Дополнения:
Пример запроса:
```
["url1", … "urlN"]
```
Пример ответа:
```
{
   url1 : server1_response_content,
   …
   urlN : serverN_response_content
}
```