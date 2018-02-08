FROM alpine:3.5
RUN mkdir -p /opt/app
ADD ./guul /opt/app
ENV TIME_ZONE=Asia/Shanghai
RUN echo "${TIME_ZONE}" > /etc/timezone \
    && ln -sf /usr/share/zoneinfo/${TIME_ZONE} /etc/localtime
WORKDIR /opt/app
RUN chmod 755 ./guul
EXPOSE 1508
CMD ["./guul"]

