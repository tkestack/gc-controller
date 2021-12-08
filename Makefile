all: push
TAG = 0.0.1
PREFIX = ccr.ccs.tencentyun.com/lutzowguo/tke-platform-gc-controller

push: container clean
	docker push $(PREFIX):$(TAG)

container: binary
	docker build -t $(PREFIX):$(TAG) .

binary: 
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-s -w' -o ./tke-platform-gc-controller cmd/platform-gc-controller.go

clean:
	rm tke-platform-gc-controller