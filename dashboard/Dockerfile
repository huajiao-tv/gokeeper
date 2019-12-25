# build api
FROM golang:1.12

COPY ./api /go/src/github.com/huajiao-tv/gokeeper/dashboard/api

WORKDIR /go/src/github.com/huajiao-tv/gokeeper/dashboard/api

RUN go install \
        -ldflags '-extldflags "-fno-PIC -static"' \
        -buildmode pie \
        -tags 'osusergo netgo static_build' \
		github.com/huajiao-tv/gokeeper/dashboard/api

# build ui
FROM node:11

COPY ./ui /data/dashboard-ui

WORKDIR /data/dashboard-ui

RUN npm install -g cnpm --registry=https://registry.npm.taobao.org
RUN rm -rf node_modules && cnpm install && cnpm run build

# image
FROM alpine:3.9

# install nginx
RUN apk update && apk add nginx && mkdir -p /run/nginx/

COPY ./conf/nginx.conf /etc/nginx/conf.d/default.conf

# dashboard api
COPY --from=0 /go/bin/api /data/dashboard/bin/api

# dashboard ui
COPY --from=1 /data/dashboard-ui/dist /var/www/dashboard


COPY ./run.sh /run.sh

EXPOSE 80

CMD ["sh", "/run.sh"]
