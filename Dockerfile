# стадия сборки
FROM golang:alpine AS build-env
ADD . /src
RUN cd /src && go build -o goapp ./test2/main.go

# финальная стадия
FROM alpine
WORKDIR /app
COPY --from=build-env /src/goapp /app/
ENTRYPOINT ./goapp




