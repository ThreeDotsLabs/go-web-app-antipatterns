FROM golang:1.16
RUN go get github.com/cespare/reflex
COPY reflex.conf /
COPY entrypoint.sh /
COPY wait-for-mysql.go /
RUN go build -o /wait-for-mysql /wait-for-mysql.go
ENTRYPOINT ["/entrypoint.sh"]
