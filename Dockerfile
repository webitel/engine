FROM node:slim
MAINTAINER Vitaly Kovalyshyn "v.kovalyshyn@webitel.com"

ENV VERSION 4.0
ENV NODE_TLS_REJECT_UNAUTHORIZED 0

COPY src /core

WORKDIR /core
RUN npm install && npm cache clear

EXPOSE 10022
ENTRYPOINT ["node", "app.js"]
