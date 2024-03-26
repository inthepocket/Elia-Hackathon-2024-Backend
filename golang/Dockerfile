FROM golang:1.22.1-alpine

COPY ./src /app/src
COPY ./go.mod ./go.sum /app/src/
WORKDIR /app/src
RUN go build -o  /app/bin/happyhour_backend 

CMD /app/bin/happyhour_backend