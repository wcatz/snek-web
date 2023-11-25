# snek-web

Get block information in realtime from IOG's public relay. Or point it to one of your own.

Requires standalone Tailwindcss binary sudo mv it to /usr/local/bin so it's available.
[Tailwind CLI](https://tailwindcss.com/blog/standalone-cli)

clone repo and cd into it.
go mod tidy

## Build the minimized CSS

tailwindcss -m -i static/css/main.css -o static/css/output.min.css --watch

## Run main.go

go run main.go

visit [http:localhost:8080](http:localhost:8080)
