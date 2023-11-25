# snek-web

Get block information in realtime from IOG's public relay. Or point it to one of your own.

## To run just

clone repo and cd into it and..

go mod tidy

go run main.go

and visit [http://localhost:8080](http://localhost:8080)

## develop

Requires standalone Tailwindcss binary. Follow the instructions. Make it executable and put it in your $PATH

Use sudo to move it to /usr/local/bin for example.

[Tailwind CLI](https://tailwindcss.com/blog/standalone-cli)

Then you can cd into the project folder and run..

tailwindcss -m -i static/css/main.css -o static/css/output.min.css --watch

To watch and restart go on changes you can use [Air](https://github.com/cosmtrek/air) in a second terminal.

