# стадия сборки
FROM golang:alpine AS build-env
ADD . /src
RUN cd /src && go build -o goapp ./cmd/main.go

# финальная стадия
FROM alpine
WORKDIR /app
COPY --from=build-env /src/goapp /app/
ENTRYPOINT ./goapp




