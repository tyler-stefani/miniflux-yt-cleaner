FROM golang:alpine as BUILD

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build

FROM alpine as RUN

WORKDIR /app
COPY --from=BUILD /app/miniflux-yt-cleaner /app/
CMD ["miniflux-yt-cleaner"]

FROM alpine as SCHEDULE

COPY --from=BUILD /app/miniflux-yt-cleaner /app/
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh
ENTRYPOINT [ "./entrypoint.sh" ]
CMD ["crond", "-f", "-d", "8"]
