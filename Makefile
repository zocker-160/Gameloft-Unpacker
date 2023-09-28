FILES=unpacker.go

lin: $(FILES)
	go build -o unpacker.bin unpacker.go
	strip unpacker.bin

win: $(FILES)
	GOOS=windows go build -ldflags="-s -w" unpacker.go

all: lin win
