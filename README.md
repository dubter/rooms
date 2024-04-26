### V2 Rooms

В данном репозитории реализована V2-версия мессенджера **Rooms**. 

Вы можете пользоваться сервисом **Rooms** уже сейчас: [rooms.servebeer.com](https://rooms.servebeer.com)

Вы также можете ознакомится с [V1-версией](https://github.com/central-university-dev/2024-spring-ab-go-hw-1-template-dubter) проекта.

Что было добавлено по сравнению прошлой версией:
- Добавлены новые компоненты бэкенда: Kafka, Redis и микросервисы `app-consumer`, `app-websocket` которые можно создавать в нескольких инстансах.
- Образы для Postgres, Redis и Kafka взяты с [bitnami/containers](https://github.com/bitnami/containers). Одно из лучших решений на рынке.
Каждый из 3х компонент является отказоустойчивым кластером c репликацией.
- Все конфиги хранятся в едином месте - папке `config`
- Появился frontend `nodejs` и `nginx` прокси для маршрутизации запросов в компоненты frontend и backend
- Захостили на [cloud.ru](https://cloud.ru) наш сервис и используя `letsencrypt` сгенерировали сертификаты для HTTPS протокола

### Архитектура V2

![](architecture/system-design-v2.png)

### Дальнейшие планы для V3 - логика
Что в планах доработать (много что):
- Разделить на микросервисы сервера, отвечающие за Websocket connection и за аутентификацию/авторизацию
- Те микросервисы, которые отвечают за аутентификацию/авторизацию оставить работать с `Postgres`, а сами сообщения хранить в `Cassandra`.

### Дальнейшие планы для V4 - инфраструктура
- После того как все, что выше сказано будет сделано, нужно будет перенести сервис с `docker-compose` на `Kubernetes`, используя helm-charts все тех же [bitnami/charts](https://github.com/bitnami/charts)
- Сделать поверх кубера весь Observability: логи используя `fluentd`, метрики - `prometheus`, дашборды - `grafana`, трейсы - `jaeger`. Лучше всего будет поставить операторы в кубере, которые будут мониторить эти ресуры. И понадобится экспорт этого всего в какую-нибудь БД.
- Перевести Nginx Load Balancer в Ingress Operator
- Важно настроить такую технологию как `AFFINITY`, чтобы поды одного микросервиса не деплоились на одной ноде
- CICD: из подходящего берем [fluxcd](https://fluxcd.io/) для CD и Github Actions для CI

### Дальнейшие планы для V5 - облако
- Используя Terraform, пишем инфраструктуру под `Managed Kubernetes` в облаке. Конвертируем `yaml` в `terraform`, используя [утилиту k2tf](https://github.com/sl1pm4t/k2tf) 
- Разделяем на Prod и Dev стэнд в облаке 

### Установка и запуск
- `make docker-local`
- Дальше ждем по логам, когда поднимется `Postgres master` - `pg-0`.
После чего применяем миграции:
```
db=pg make migrate-up        # Создаём таблицы и индексы для postgres
make create-kafka-topic-local  # Создаём топик в кафке
```

- Стучимся в localhost:8080 по эндпоинтам:
```
POST /user/register          # Регистрация
POST /user/login             # Аутентификация
POST /user/refresh           # Эндпоинт для фронтенда для обновления JWT токенов
POST /chat/rooms             # Создание Room
GET /chat/rooms              # Получение списка всех Room
GET /chat/rooms/{id}/clients # Получение списка всех подключенных клиентов
WS /chat/rooms/{id}          # Подключение к выбранной Room
```
- Наслаждаемся) Приятнее всего использовать `Postman` в качестве клиента сервиса. В папке `tests/postman` необходимая для тестов коллекция. 

### Брал вдохновения из источников:
- [10 минутное видео на ютубе](https://www.youtube.com/watch?v=xyLO8ZAk2KE)
- 12 глава книги [System design Алекс Сюй](https://www.piter.com/collection/programmirovanie-osnovy-i-algoritmy/product/system-design-podgotovka-k-slozhnomu-intervyu)
- [видео подлиннее](https://www.youtube.com/watch?v=vvhC64hQZMk)