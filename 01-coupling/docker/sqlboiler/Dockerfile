FROM golang:1.16

RUN go install github.com/volatiletech/sqlboiler/v4@latest
RUN go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-mysql@latest

COPY wait-for-mysql.go /
RUN go build -o /wait-for-mysql /wait-for-mysql.go

COPY ./entrypoint.sh /entrypoint.sh

CMD ["/entrypoint.sh"]
