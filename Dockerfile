FROM centos
COPY tdcc-platform-gc ./
RUN mkdir -p /etc/config/
ENTRYPOINT ["./tdcc-platform-gc"]