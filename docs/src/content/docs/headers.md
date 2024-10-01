---
title: Request Headers
tableOfContents: false

---
Request Headers can either be set when creating a new client or when making a request. 


:::note
Headers set when creating a client will be used for all requests made by that client.

:::

### Setting Headers when creating a client

```go {2-3}
httpc := httpr.NewClient(
  httpr.Header("Content-Type", "application/json"),
  httpr.Header("Accept", "application/json"),
)

resp, err := httpc.Get(context.Background(), "https://hehe.gov") // Content-Type and Accept headers will be set
```