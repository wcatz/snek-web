# snek-web

Requires standalone Tailwindcss binary sudo mv it to /usr/local/bin so it's available.
[Tailwind CLI](https://tailwindcss.com/blog/standalone-cli)

clone repo and cd into it.
go mod tidy

# Start Tailwind
tailwindcss -m -i static/css/main.css -o static/css/output.min.css --watch

in another terminal run main.go

go run main.go

visit http:localhost:8080
