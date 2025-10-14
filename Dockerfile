FROM golang:latest

RUN mkdir /app
ADD . /app/
WORKDIR /app/
RUN go build -o main .
WORKDIR /app/
CMD ["/app/main", "start"]