FROM golang:1.17-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./config/ /app/config
COPY ./controllers/ /app/controllers
COPY ./middlewares/ /app/middlewares
COPY ./models/ /app/models
COPY ./server/ /app/server
COPY ./services/ /app/services
COPY ./main.go /app/

RUN apk --no-cache --update add build-base

RUN go build -o /indiescrobble

COPY ./static /app/static
COPY templates /app/templates

EXPOSE 8081

CMD [ "/indiescrobble" ]