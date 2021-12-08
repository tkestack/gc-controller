FROM centos
COPY tke-platform-gc-controller ./
RUN mkdir -p /etc/config/
ENTRYPOINT ["./tke-platform-gc-controller"]