FROM golang:1.12.0-alpine3.9

VOLUME /srv
EXPOSE 8090

WORKDIR /app
ADD . /app
RUN go build -o main .

ENTRYPOINT ["/app/main"]