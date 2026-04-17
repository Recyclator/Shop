.PHONY: dev build clean

dev:
	@echo "▶ Starting dev server..."
	templ generate --watch & termux-chroot air

build:
	@echo "▶ Building..."
	templ generate
	go build -o bin/server ./cmd/server

clean:
	@echo "▶ Cleaning generated files..."
	rm -f views/**/*_templ.go
	rm -f bin/server
