####
# 编译镜像
####
FROM golang

# 创建代码编译目录
RUN mkdir -p /go/src/github.com/huajiao-tv/gokeeper

# 设置工作目录
WORKDIR /go/src/github.com/huajiao-tv/gokeeper

# 拷贝代码
COPY ./ /go/src/github.com/huajiao-tv/gokeeper


# 编译
RUN go build

# 运行镜像
FROM alpine:latest

# 安装基本软件包
RUN apk --no-cache add ca-certificates librrd

# 运行必须的动态库软链
RUN mkdir -p /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2 && ln -s /usr/lib/librrd.so.8 /usr/lib/librrd.so.4

RUN mkdir -p /data/keeper/

WORKDIR /data/keeper/

COPY --from=0 /go/src/github.com/huajiao-tv/gokeeper/gokeeper .

COPY ./deploy/docker_compose/keeper.conf /tmp/keeper.conf

CMD ["./gokeeper", "-f", "/tmp/keeper.conf"]
#ENTRYPOINT ["tail", "-f", "/dev/null"]
