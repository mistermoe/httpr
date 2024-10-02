---
title: Query Params
tableOfContents: true
slug: 1.0/query-params
---

Query Params can be set per request:

### Per Request

```go {4-5}
httpc := httpr.NewClient()

resp, err := httpc.Get(context.Background(), "https://hehe.gov", 
  httpr.QueryParam("page", "3"),
  httpr.QueryParam("limit", "10"),
)
```

:::note
Setting the same query param multiple times is allowed per [RFC 3986](https://datatracker.ietf.org/doc/html/rfc3986#section-3.4)
:::
