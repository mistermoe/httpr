---
title: Response Bodies
tableOfContents: true
---

Response body parsers can be provided as request options. `httpr` comes with built-in helpers for parsing JSON, string, and byte slice response bodies.

## JSON

JSON response bodies can be handled using the `ResponseBodyJSON` helper function. This function allows you to specify separate structs for successful and error responses.

:::note
if the response status code is >= 400, the error response body will be parsed into the provided error struct if one is provided.
:::

```go
httpc := httpr.NewClient()

type SuccessResponse struct {
    UserID int    `json:"userId"`
    ID     int    `json:"id"`
    Title  string `json:"title"`
    Body   string `json:"body"`
}

type ErrorResponse struct {
    Message string                  `json:"message"`
    Code    optional.Option[int]    `json:"code"`
    Field   optional.Option[string] `json:"field"`
}

var successBody SuccessResponse
var errBody ErrorResponse

resp, err := httpc.Get(
    context.Background(),
    "https://api.example.com/posts/1",
    httpr.ResponseBodyJSON(&successBody, &errBody),
)

if err != nil {
    // Handle error
}

if resp.StatusCode >= 400 {
    // Handle error response using errBody
} else {
    // Process successful response using successBody
}
```

## String

String response bodies can be handled using the `ResponseBodyString` helper function.

```go
httpc := httpr.NewClient()

var responseBody string

resp, err := httpc.Get(
    context.Background(),
    "https://api.example.com/text",
    httpr.ResponseBodyString(&responseBody),
)

if err != nil {
    // Handle error
}

fmt.Println("Response:", responseBody)
```

## Bytes

Byte slice response bodies can be handled using the `ResponseBodyBytes` helper function.

```go
httpc := httpr.NewClient()

var responseBody []byte

resp, err := httpc.Get(
    context.Background(),
    "https://api.example.com/binary",
    httpr.ResponseBodyBytes(&responseBody),
)

if err != nil {
    // Handle error
}

// Process the byte slice responseBody
```

## Custom Response Body Handler

You can create a custom response body handler using the `ResponseBody` function.

```go
httpc := httpr.NewClient()

customHandler := httpr.ResponseBody(func(resp *http.Response) error {
    // Your custom logic here
    // Read the response body, process it, etc.
    return nil
})

resp, err := httpc.Get(
    context.Background(),
    "https://api.example.com/custom",
    customHandler,
)

if err != nil {
    // Handle error
}
```

