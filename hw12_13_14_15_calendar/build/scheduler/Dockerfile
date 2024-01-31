# Собираем в гошке
FROM golang:1.20 as build

ENV BIN_FILE /opt/calendar/calendar-scheduler
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

# Собираем статический бинарник Go (без зависимостей на Си API),
# иначе он не будет работать в alpine образе.
ARG LDFLAGS
RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} cmd/scheduler/*

# На выходе тонкий образ
FROM alpine:3.9

LABEL ORGANIZATION="OTUS Online Education"
LABEL SERVICE="calendar-scheduler"
LABEL MAINTAINERS="student@otus.ru"

ENV BIN_FILE "/opt/calendar/calendar-scheduler"
COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENV CONFIG_FILE /etc/calendar/scheduler_config.toml
COPY ./configs/scheduler_config.toml ${CONFIG_FILE}

ENV CONFIG_MIGRATION_DIR /etc/calendar/migrations
COPY ./migrations ${CONFIG_MIGRATION_DIR}

CMD ${BIN_FILE} -config ${CONFIG_FILE}
