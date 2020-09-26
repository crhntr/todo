# todo
A simple todo app using Go+wasm.

The "main" package for the server is in the `cmd/todo-server` directory.

The "main" package for the webapp is in the `pages/todo` directory.
Note the proximity of the index.html and main.go. I think this makes it easier to refactor.

The app can be pushed to dokku and uses the Dockerfile type deployment. [Ensure your exposed ports are configured properly](http://dokku.viewdocs.io/dokku/deployment/methods/dockerfiles/).

## Development setup

To run locally execute

`make setup && make`
