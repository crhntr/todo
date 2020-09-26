FROM golang as build
WORKDIR /go/src/github.com/crhntr/todo/
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o todo-server ./cmd/todo-server
RUN GOOS=js GOARCH=wasm go build -o todo.wasm ./pages/todo

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN mkdir -p bin assets/wasm assets/scripts
COPY assets assets
COPY pages pages
COPY --from=todo-server /go/src/github.com/crhntr/todo/todo-server ./bin/todo-server
COPY --from=build /go/src/github.com/crhntr/todo/todo.wasm ./assets/wasm/todo.wasm
COPY --from=build /usr/local/go/misc/wasm/wasm_exec.js ./assets/scripts
EXPOSE 80
ENV PORT=80
ENTRYPOINT ["./bin/todo-server"]