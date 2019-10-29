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
	GOOS=$(os) GOARCH=$(arch) go build -o 'bin/$(longname)/$(name)' .
	cd bin/$(longname) && zip $(longname).zip $(name)

clean:
	rm -rvf bin/*

.PHONY: checkenv release $(PLATFORMS) clean
