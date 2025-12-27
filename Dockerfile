# compile

FROM docker.io/golang:alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY --exclude=.git ./ ./
COPY .git/refs/heads/main ./HEAD

# ARG GIT_COMMIT # cant run git in quadlet build

RUN \
GOEXPERIMENT=greenteagc \
CGO_ENABLED=0 GOOS=linux \
go build -ldflags="-s -w \
-X 'github.com/makinori/mikogo/env.GIT_COMMIT=$(cat HEAD | head -c 8)'\
" -o mikogo

# create image

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=build /app/mikogo /

CMD ["/mikogo"]