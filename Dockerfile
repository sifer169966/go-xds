# Defining App builder image
FROM public.ecr.aws/docker/library/golang:1.22.3-alpine3.19 AS builder
RUN apk update; \
    apk add --no-cache \
    git
WORKDIR /app
ENV GO111MODULE=on
ARG NETRC_USER
ARG NETRC_TOKEN
RUN echo -e "machine gitlab.com\nlogin $NETRC_USER\npassword $NETRC_TOKEN\n" > ~/.netrc
RUN chmod 600 ~/.netrc
# RUN go env -w GOPRIVATE=github.com/*
# RUN go env -w GOSUMDB=off
RUN go install -v -trimpath github.com/grpc-ecosystem/grpc-health-probe@v0.4.26
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 go build -v -trimpath -o .bin/xdsserver main.go
RUN rm ~/.netrc

FROM public.ecr.aws/docker/library/alpine:3.19 as release
RUN apk add --no-cache --update ca-certificates curl
RUN adduser --disabled-password --gecos "" --shell "/sbin/nologin" --home "/nonexistent" --no-create-home --uid 10014 "app"
COPY --from=builder /go/bin/grpc-health-probe /usr/bin/
COPY --from=builder /app/.bin/xdsserver /app/cmd/
USER 10014
WORKDIR /app
COPY ./healthz.sh ./healthz.sh
CMD ["cmd/xdsserver"]
