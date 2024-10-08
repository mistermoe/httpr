---
title: Request Bodies
tableOfContents: true
slug: 1.0/request-body
---

Request bodies can be set when making a request. The HTTP client supports various types of request bodies.

## JSON

JSON request bodies can be set using the `RequestBodyJSON` helper function which will `json.Marshal` the provided value.

:::note
The `Content-Type` header will be set to `application/json` when using `RequestBodyJSON`.
:::

### Strongly Typed

```go
httpc := httpr.NewClient()

type Post struct {
    UserID int    `json:"userId"`
    ID     int    `json:"id"`
    Title  string `json:"title"`
    Body   string `json:"body"`
}

post := Post{
    UserID: 1,
    ID:     1,
    Title:  "foo",
    Body:   "bar",
}

resp, err := httpc.Post(
    context.Background(),
    "https://jsonplaceholder.typicode.com/posts",
    httpr.RequestBodyJSON(post),
)
```

### Map

```go
httpc := httpr.NewClient()

data := map[string]any{
    "userId": 1,
    "id":     1,
    "title":  "foo",
    "body":   "bar",
}

resp, err := httpc.Post(
    context.Background(),
    "https://jsonplaceholder.typicode.com/posts",
    httpr.RequestBodyJSON(post),
)
```

## String

String request bodies can be set using the `RequestBodyString` helper function.

:::note
The `Content-Type` header will be set to `text/plain` when using `RequestBodyString`.
:::

```go
httpc := httpr.NewClient()

resp, err := httpc.Post(context.Background(), "https://api.example.com/text",
    httpr.RequestBodyString("Hello, World!"),
)
```

## Form Data

Form data request bodies can be set using the `RequestBodyForm` helper function.

:::note
The `Content-Type` header will be set to `application/x-www-form-urlencoded` when using `RequestBodyForm`.
:::

```go
httpc := httpr.NewClient()

formData := url.Values{}
formData.Add("username", "johndoe")
formData.Add("password", "secret")

resp, err := httpc.Post(context.Background(), "https://api.example.com/form",
    httpr.RequestBodyForm(formData),
)
```

## Bytes

Bytes request bodies can be set using the `RequestBodyBytes` helper function.

```go
httpc := httpr.NewClient()

myBytes := []byte{0x00, 0x01, 0x02, 0x03}
resp, err := httpc.Post(context.Background(), "https://api.example.com/binary",
    httpr.RequestBodyBytes("application/octet-stream", myBytes),
)
```

## Streaming Request Body

If you need to send a large amount of data that you don't want to load into memory all at once, use the `RequestBodyStream` helper function.

```go
httpc := httpr.NewClient()

fileReader, err := os.Open("large_file.dat")
if err != nil {
    // Handle error
}
defer fileReader.Close()

resp, err := httpc.Post(context.Background(), "https://api.example.com/upload",
    httpr.RequestBodyStream("application/octet-stream", fileReader),
)
```

:::caution
This does not yet work with `httpr.Inspect`
:::

## Custom Request Body Helper

```go
httpc := httpr.NewClient()

customBody := httpr.RequestBody("application/custom", func() (io.Reader, error) {
    // Your custom logic here
    return strings.NewReader("Custom data"), nil
})

resp, err := httpc.Post(context.Background(), "https://api.example.com/custom",
    customBody,
)
```
