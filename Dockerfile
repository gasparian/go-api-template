FROM golang:1.22 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .
# NOTE: statically linked executable for x86 linux machine
RUN GOOS=linux GOARCH=amd64 make static

FROM scratch
COPY --from=build /app/bin/server /server
COPY --from=build /app/configs/*.toml /configs/
CMD ["/server"]