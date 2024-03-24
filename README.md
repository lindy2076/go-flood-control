# флуд-контроль с помощью redis timeseries 

Т.к. несколько экземпляров приложения могут стучаться к флуд-контролю, то нужно использовать какую-нибудь БД.

Первая идея: использовать SQL. Проблема: много возни с индексами по `timestamp`, т.к. обновление данных очень частое, следовательно перестройка
индексов будет занимать много времени. SQL отбросил.

Вторая идея: использовать redis. Проблема: я не работал с redis до этого момента. Посмотрел, какие данные можно хранить. Увидел timeseries. Решил воспользоваться ввиду наличия всех нужных методов: добавление timestamp по ключу (очевидно `userID`), а также любого числового значения (количество запросов во время `timestamp`); плюс наличие агрегации, автоматического удаления старых меток (если древнее новейшей минус `retention` секунд, а нам как раз нужно хранить только за последние `retention` секунд) и готовой concurrency - если в один `timestamp` появится несколько нажатий, то они складываются. Осталось лишь написать это на go.

Есть готовый пакет для работы с redis - `github.com/redis/go-redis`, его и использовал. В переменные окружения нужно добавить следующие переменные: 
```
FC_REDIS_HOST=localhost
FC_REDIS_PORT=6379
FC_REDIS_PASSWORD=strongpassword
FC_RETENTION=5
FC_MAXCHECKS=4
```
где `FC_RETENTION` и `FC_MAXCHECKS` вообще говоря не обязательны, их можно прописать прямо в коде. `FC_RETENTION` - количество секунд, за которые можно сделать не более `FC_MAXCHECKS` запросов.

В коде:

```go
// создать клиент редиса
redisClient := redis.NewClient(&redis.Options{
    Addr:     fmt.Sprintf("%s:%s", FC_REDIS_HOST, FC_REDIS_PORT),
    Password: FC_REDIS_PASSWORD,
    DB:       0,
})
// задать контекст
ctx := context.Background()
// получить собственно флуд-контроль
var rfc FloodControl = &fc.RedisFloodController{
    Client:           redisClient,
    RetentionSeconds: FC_RETENTION,
    MaxChecks:        FC_MAXCHECKS,
}
// проверить доступность
v, err := rfc.Check(ctx, 1234567890)

```

В репе лежит простой `docker-compose.yml` для запуска redis с timeseries.

RedisFloodController.Check аккуратно возвращает ошибки, не роняя приложение, что вопщем то важно в проде.

#### супер быстрый старт
```
cp .env.sample .env
export $(xargs < .env)
docker compose -f docker-compose.yml up -d
go run .
```

//Есть тесты. Гонка проходит, по крайней я запсукал тесты 10 раз подряд с очисткой кеша. Для тестов используется боевой редис, в miniredis будто бы ещё нет поддержки ts.