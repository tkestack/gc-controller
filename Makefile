all: push
TAG = 1.0.0
PREFIX = ccr.ccs.tencentyun.com/lutzowguo/tdcc-platform-gc

push: container clean
	docker push $(PREFIX):$(TAG)

container: binary
	docker build -t $(PREFIX):$(TAG) .

binary: 
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-s -w' -o ./tdcc-platform-gc cmd/platform-gc-controller.go

clean:
	rm tdcc-platform-gc