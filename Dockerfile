FROM golang:latest AS build

# create binary
ADD . /build
WORKDIR /build
RUN go version
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /opmbuilder .

FROM gruebel/upx:latest as upx
COPY --from=build /opmbuilder /opmbuilder
RUN upx --best --lzma /opmbuilder

FROM alpine:latest
COPY --from=upx /opmbuilder /opmbuilder
COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
