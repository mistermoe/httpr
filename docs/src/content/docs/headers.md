---
title: Request Headers
tableOfContents: true

---
Request Headers can either be set when creating a new client or when making a request. 

### Global

```go {2-3}
httpc := httpr.NewClient(
  httpr.Header("Authorization", "some auth token"), // Set the Authorization header for all requests
  httpr.Header("Accept", "application/json"), // Set the Accept header for all requests
)

resp, err := httpc.Get(context.Background(), "https://hehe.gov")
```

:::note
Headers set when creating a client will be used for all requests made by that client.

:::

### Per Request

```go {4-5}
httpc := httpr.NewClient()

resp, err := httpc.Get(context.Background(), "https://hehe.gov", 
  httpr.Header("X-Request-ID", "1234"),
  httpr.Header("Accept", "application/json"),
)
```

:::note
if headers are set both globally and per request, they will be merged with the per request headers taking precedence.
:::