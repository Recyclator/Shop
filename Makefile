.PHONY: dev build clean

dev:
	@echo "▶ Starting dev server..."
	npx @tailwindcss/cli -i ./static/css/input.css -o ./static/css/app-ui.css --watch & templ generate --watch & air

build:
	@echo "▶ Building..."
	npx @tailwindcss/cli -i ./static/css/input.css -o ./static/css/app-ui.css
	templ generate
	go build -o bin/server .

clean:
	@echo "▶ Cleaning generated files..."
	rm -f views/**/*_templ.go
	rm -f bin/server
