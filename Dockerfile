# build stage
FROM golang:1.12.4-alpine3.9 AS build-env
RUN apk add git
COPY . /src
RUN cd /src && go build -o docker-authorizer

# final stage
FROM alpine:3.9
WORKDIR /app
COPY --from=build-env /src/docker-authorizer /app/
CMD /app/docker-authorizer
