FILES=unpacker.go

lin: $(FILES)
	go build -o unpacker.linux.bin unpacker.go
	strip unpacker.linux.bin

win: $(FILES)
	GOOS=windows go build -ldflags="-s -w" unpacker.go

all: lin win
