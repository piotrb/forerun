# credit: https://vic.demuzere.be/articles/golang-makefile-crosscompile/
PLATFORMS := darwin/386 darwin/amd64 linux/386 linux/amd64

ifndef TAG
	TAG = ${shell git describe}
endif

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))
name = forerun
longname = $(name)-$(TAG)-$(os)-$(arch)

release: clean $(PLATFORMS)

$(PLATFORMS):
	GOOS=$(os) GOARCH=$(arch) go build -ldflags "-X main.version=$(TAG)" -o 'bin/$(longname)/$(name)' .
	cd bin/$(longname) && zip $(longname).zip $(name)

run:
	go run -ldflags "-X main.version=$(TAG)" . ${step}

run-docker:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(TAG)" -o 'test/forerun' .
	docker build -t forerun:latest test/
	docker run --name forerun -it --rm forerun:latest /usr/bin/forerun ${step}

exec-docker:
	docker exec -it forerun bash -ls

kill-docker:
	docker kill forerun

clean:
	rm -rvf bin/*

.PHONY: checkenv release $(PLATFORMS) clean run run-docker exec-docker kill-docker
