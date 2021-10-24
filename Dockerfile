FROM golang:1.17-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o bin/icc

FROM alpine
WORKDIR /root/
COPY --from=builder /app/bin/icc /usr/local/bin/
CMD [ "icc" ]
ENV AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
ENV AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}