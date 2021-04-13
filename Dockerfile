FROM golang:1.14-alpine as builder
WORKDIR /usr/src/app
ENV GOPROXY=https://goproxy.cn
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
  apk add --no-cache upx ca-certificates tzdata
COPY ./go.mod ./
COPY ./go.sum ./
RUN go mod download
COPY . .
#RUN go get  github.com/containerd/containerd
RUN  CGO_ENABLED=0 go build -o server  -ldflags "-X 'github.com/prometheus/common/version.BuildUser=root@n9e'  -X 'github.com/prometheus/common/version.BuildDate=`date`' -X 'github.com/prometheus/common/version.Version=`cat VERSION`'"
#FROM scratch as runner
#FROM busybox  as runner
FROM yauritux/busybox-curl  as runner
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/src/app/server /opt/app/k8s-mon
ENTRYPOINT [ "/opt/app/k8s-mon" ]
CMD [ "--config.file=/etc/k8s-mon.yml"]

