FROM golang:1.22.1-alpine

COPY ./src/go.* /app/src/
COPY ./precompile/main.go /app/src/
WORKDIR /app/src
RUN go build -v -o /app/bin/dummy

COPY ./src/*.go /app/src/
COPY ./.env /app/bin/
RUN go build -v -o /app/bin/happyhour_backend 

CMD /app/bin/happyhour_backend