FROM golang:1.22
RUN go install github.com/cespare/reflex@latest
COPY reflex.conf /
COPY entrypoint.sh /
COPY wait-for-postgres.go /
RUN go build -o /wait-for-postgres /wait-for-postgres.go
ENTRYPOINT ["/entrypoint.sh"]
