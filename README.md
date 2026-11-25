# Go и базы данных: ORM vs SQL

Учебный репозиторий для вебинара. Требуются Go 1.25, Docker Compose v2, make, интернет для первоначальной установки зависимостей. База локальная, данные синтетические. Не используйте production DSN: демонстрации изменяют данные.

```sh
make deps
make up
make migrate
make seed
make test
make demo-crud
make demo-n-plus-one
make demo-zero-values
make demo-select-star
make demo-filters
make demo-transactions
make demo-bulk
make plans
make bench
make profile
go tool pprof -http=:8080 cpu.out
make down
```

`make deps` получает актуальные версии на момент запуска и генерирует sqlc; затем сохраните go.mod/go.sum в Git для воспроизводимости. Если зависимости требуют более нового Go, закрепите совместимые версии перед вебинаром. `make seed` удаляет предыдущие данные и создаёт миллион заказов; повторять перед сравнениями при необходимости. Денежные суммы хранятся в копейках. SQLX и GORM используют общий database/sql пул, pgx и sqlc — общий pgxpool. Бенчмарки выполняются последовательно, с одинаковой проекцией; цифры не предопределены.

## Сценарий

| Команда | Что наблюдать |
|---|---|
| demo-crud | Одинаковые данные четырёх реализаций |
| demo-n-plus-one | 101 SQL против 2 |
| demo-zero-values | Updates со структурой пропускает ноль |
| demo-select-star | Влияние ширины строки; грубое измерение, не бенчмарк |
| demo-filters | Squirrel и параметры |
| demo-transactions | Lost update против атомарного UPDATE; результаты гонки варьируются |
| demo-bulk | GORM batch против COPY, включая различие RETURNING |
| plans | План запроса до и после индекса |
| bench | ns/op, B/op, allocs/op |

Полный материал: docs/webinar.md. Этот starter не включает HTTP API, Testcontainers и мониторинг; интеграционные тесты используют Compose. Код отформатирован, но требует проверки сборки и запуска в среде Go 1.25 с PostgreSQL.
# go-orm-vs-sql
