# `httpr`

try-hard general purpose http client for golang

> [!WARNING]
> 🚧 WIP 👷


## Features
* BaseURL configuration - set base url for all requests
* Setting custom headers both globally and per request
* Setting custom query params
* Supplying strongly typed request bodies 
* Unmarshalling response bodies into strong types (for success and error responses)
* Interceptor support for request/response modification and inspection
* Built-in request/response inspector for debugging
* Instrumented with metrics and traces for observability