FROM golang:1.22.1-alpine

COPY ./golang/src /app/src
COPY ./golang/go.mod ./golang/go.sum /app/src/
WORKDIR /app/src
RUN go build -o  /app/bin/happyhour_backend 

CMD /app/bin/happyhour_backend