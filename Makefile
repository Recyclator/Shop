export PATH := $(PATH):$(HOME)/go/bin
export INITIAL_ADMIN_PASSWORD ?= admin123
export JWT_SECRET ?= dev-secret-key-change-in-production-123456

.PHONY: dev build clean

dev:
	@echo "▶ Starting dev server..."
	node node_modules/.bin/tailwindcss -i ./static/css/input.css -o ./static/css/app-ui.css --watch & templ generate --watch & air

build:
	@echo "▶ Building..."
	node node_modules/.bin/tailwindcss -i ./static/css/input.css -o ./static/css/app-ui.css
	templ generate
	go build -o bin/server .

clean:
	@echo "▶ Cleaning generated files..."
	rm -f views/**/*_templ.go
	rm -f bin/server
