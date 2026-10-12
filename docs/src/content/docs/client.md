---
title: Client
tableOfContents: true

---
A client is created with `httpr.NewClient`, configured with options. Options that apply to every request, such as a base URL or headers, are set here.

```go
httpc := httpr.NewClient(
  httpr.BaseURL("https://someapi.io"),
  httpr.Timeout(10 * time.Second),
)
```

### Sending Requests

The client has a method per HTTP method. Each takes a context, a URL or path, and any request options.

```go
resp, err := httpc.Get(ctx, "/users")
resp, err := httpc.Post(ctx, "/users", httpr.RequestBodyJSON(user))
resp, err := httpc.Put(ctx, "/users/1", httpr.RequestBodyJSON(user))
resp, err := httpc.Patch(ctx, "/users/1", httpr.RequestBodyJSON(changes))
resp, err := httpc.Delete(ctx, "/users/1")
```

`SendRequest` takes the method as a string, for any method without its own function:

```go
resp, err := httpc.SendRequest(ctx, http.MethodHead, "/users/1")
```

:::note
A response with a 4xx or 5xx status is not an error. Check `resp.StatusCode`, or use a [response body](../response-body/) option to unmarshal error responses. Unless a response body option reads it, close `resp.Body` when you're done with it.
:::

### Base URL

```go {2}
httpc := httpr.NewClient(
  httpr.BaseURL("https://someapi.io"),
)

resp, err := httpc.Get(ctx, "/users") // https://someapi.io/users
```

A URL that starts with `http://` or `https://` is used as is, and the base URL is ignored for that request:

```go
resp, err := httpc.Get(ctx, "https://otherapi.io/status")
```

### Custom http.Client

`HTTPClient` sets the `http.Client` that sends requests, for example to use a custom transport:

```go {2}
httpc := httpr.NewClient(
  httpr.HTTPClient(http.Client{Transport: myTransport}),
)
```

Without it, the client uses a zero `http.Client{}`: no timeout, and up to 10 redirects followed.

### Timeout

`Timeout` limits how long a request can take, including reading the response body:

```go {2}
httpc := httpr.NewClient(
  httpr.Timeout(10 * time.Second),
)
```

### Redirects

By default the client follows redirects. `NoRedirects` makes it return a redirect (3xx) response as is instead:

```go {2}
httpc := httpr.NewClient(
  httpr.NoRedirects(),
)

resp, err := httpc.Post(ctx, "https://hehe.gov/webhooks", httpr.RequestBodyJSON(event))
// resp.StatusCode is 302 if the server redirected, and the redirect was not followed
```

:::caution
`HTTPClient` replaces the client's `http.Client`, while `Timeout` and `NoRedirects` change it. Pass `HTTPClient` first, or it discards them.
:::
