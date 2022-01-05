FROM golang:1.17
#RUN apt-get update
#RUN apt-get install curl net-tools telnet -y
#RUN apk add --update curl
#RUN echo "hosts: files dns" > /etc/nsswitch.conf
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN mkdir /root/bin
RUN mkdir -p /etc/config/

COPY tdcc-platform-gc /root/bin/tdcc-platform-gc
WORKDIR /root
ENTRYPOINT ["/root/bin/tdcc-platform-gc"]