FROM golang:1.16-alpine as builder

ARG test
ENV TEST=$test
RUN echo "$TEST"

WORKDIR /src

ADD . .
RUN go build -o /app pkg/main.go

FROM alpine
ARG commit_hash
ENV COMMIT_HASH=$commit_hash

RUN apk add --no-cache curl ca-certificates

COPY --from=builder /app /

# Bind the app to 0.0.0.0 so it can be seen from outside the container
ENV ADDR=0.0.0.0

EXPOSE 3000

CMD echo "Starting with [$COMMIT_HASH]" && /app