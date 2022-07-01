############################################################ 
# Dockerfile to build golang Installed Containers 

# Based on alpine

############################################################

FROM golang:1.18.0 AS builder

COPY . /src
WORKDIR /src

RUN GOPROXY="https://goproxy.cn,direct" make build

FROM alpine:3.13

RUN mkdir /keel
COPY --from=builder /src/dist/linux_amd64/release/tkeel-device /keel


WORKDIR /keel
CMD ["/keel/tkeel-device"]