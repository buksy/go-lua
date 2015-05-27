all:
	GOPATH=`pwd` go build -gcflags "-N -l" lua &&\
	GOPATH=`pwd` go build --gcflags "-N -l" src/test/test-lua.go

run: all
	./test-lua
