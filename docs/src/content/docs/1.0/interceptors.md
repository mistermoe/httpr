---
title: Interceptors
tableOfContents: true
slug: 1.0/interceptors
---

Interceptors can be used to provide custom logic that runs *prior* to sending the request and/or *after* receiving a response. Use-cases for interceptors include:

* logging
* observabilitiy
* retries
* circuit breaking.

Specifically, an interceptor is a function that wraps the logic for making an HTTP request. Interceptors can be used to modify the request before it's sent, modify the response before it's returned, or to perform some action before or after the request is made.

## Creating an Interceptor

Interceptors can be implemented by writing a function that conforms to the `HandleFunc` type or by creating a struct that implements the `Interceptor` interface.

### `HandleFunc` Approach

`HandleFunc` is a function type that matches the signature of the `Handle` method in the `Interceptor` interface. This approach is ideal for simple interceptors or when you don't need to maintain state between requests.

```go
type HandleFunc func(ctx context.Context, req *http.Request, next Interceptor) (*http.Response, error)
```

Here's an example of a simple interceptor that adds an `Authorization` header that's computed based on attributes of the request (e.g. path, request body etc.):

```go
auth := httpr.HandleFunc(func(ctx context.Context, req *http.Request, next httpr.Interceptor) (*http.Response, error) {
    // Compute the signature
    signature := computeSignature(req)

    // Add the Authorization header
    req.Header.Set("Authorization", signature)

    // Call the next interceptor in the chain
    return next.Handle(ctx, req, next)
})

client := httpr.NewClient(httpr.Intercept(auth))
```

### `Interceptor` Interface Approach

The `Interceptor` interface is more suitable when you need to maintain state or when you're creating a more complex interceptor that might benefit from being a struct with methods.

```go
type Interceptor interface {
    Handle(ctx context.Context, req *http.Request, next Interceptor) (*http.Response, error)
}
```

Here's an example of a logging interceptor that logs the request method and URL before making the request:

```go
package main

import (
  "context"
  "fmt"
  "io"
  "log"
  "net/http"
  "os"

  "github.com/mistermoe/httpr"
)

type RequestLogger struct {
  Logger *log.Logger
}

func (l *RequestLogger) Handle(ctx context.Context, req *http.Request, next httpr.Interceptor) (*http.Response, error) {
  l.Logger.Printf("Sending %s request to %s\n", req.Method, req.URL)
  return next.Handle(ctx, req, next)
}

func main() {
  logger := &RequestLogger{Logger: log.New(os.Stdout, "", 0)}

  client := httpr.NewClient(httpr.Intercept(logger))

  resp, err := client.Get(context.Background(), "https://httpbin.org/get")
  if err != nil {
    fmt.Printf("Error: %v\n", err) //nolint:forbidigo // example
    return
  }
  defer resp.Body.Close()

  body, _ := io.ReadAll(resp.Body)
  fmt.Printf("Response: %s\n", body) //nolint:forbidigo // example
}
```

:::tip
`httpr` provides a handful of built-in interceptors that you can use out of the box. Check out the [Observability](/1.0/observability) and [Request Inspection](/1.0/inspect). Worth having a look at how they've been implemented if you're looking to create your own. Source code is available [here](https://github.com/mistermoe/httpr/blob/main/observer.go) and [here](https://github.com/mistermoe/httpr/blob/main/inspect.go) respectively
:::

## Using Interceptors

Interceptors can be provided to the client using the `Intercept` option. You can provide multiple interceptors to the client like so:

```go
// assume auth and logger are previously defined interceptors

httpc := httpr.NewClient(
    httpr.Intercept(auth),
    httpr.Intercept(logger),
)
```

:::note
Interceptors are executed in the order they are provided to the client. The first interceptor provided will be the first to run and the last interceptor provided will be the last to run.
:::

Interceptors can also be provided to individual requests using the `Intercept` option:

```go
// assume auth is a previously defined interceptor

httpc := httpr.NewClient()
httpr.Get(context.Background(), "https://api.example.com", httpr.Intercept(auth))
```

:::note
Interceptors provided at the client level will run for every request made by the client. Interceptors provided at the request level will only run for that specific request. client level interceptors will run before request level interceptors.
:::
