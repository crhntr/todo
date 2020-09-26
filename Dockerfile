FROM golang as todo-webapp
WORKDIR /go/src/github.com/crhntr/todo/
COPY . .
RUN go mod download
RUN GOOS=js GOARCH=wasm go build -o todo-webapp.wasm ./pages/todo

FROM golang as todo-server
WORKDIR /go/src/github.com/crhntr/todo/
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o todo-server ./cmd/todo-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN mkdir -p bin assets/wasm assets/scripts
COPY assets assets
COPY pages pages
COPY --from=todo-server /go/src/github.com/crhntr/todo/todo-server ./bin/todo-server
COPY --from=todo-webapp /go/src/github.com/crhntr/todo/todo-webapp.wasm ./assets/wasm/todo-webapp.wasm
COPY --from=todo-webapp /usr/local/go/misc/wasm/wasm_exec.js ./assets/scripts
EXPOSE 80
ENV PORT=80
ENTRYPOINT ["./bin/todo-server"]