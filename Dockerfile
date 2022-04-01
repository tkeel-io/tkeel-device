#
# Qingcloud tkeel-device Dockerfile
#
#FROM alpine:3.13
FROM alpine:3.13
RUN mkdir /keel
ADD dist/linux_amd64/release/tkeel-device /keel
#ADD config.yml /keel
WORKDIR /keel
CMD ["/keel/tkeel-device"]