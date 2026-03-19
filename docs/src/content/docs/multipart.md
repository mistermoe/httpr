---
title: Multipart Requests
tableOfContents: true
---

Multipart request bodies (`multipart/form-data`) can be constructed declaratively using `RequestBodyMultipart` with composable parts.

:::note
The `Content-Type` header will be set to `multipart/form-data` with an auto-generated boundary when using `RequestBodyMultipart`.
:::

## Form Fields

Use `FormField` to add text fields to a multipart request.

```go
httpc := httpr.NewClient()

resp, err := httpc.Post(
    context.Background(),
    "https://api.example.com/submit",
    httpr.RequestBodyMultipart(
        httpr.FormField("username", "moe"),
        httpr.FormField("description", "a cool thing"),
    ),
)
```

## File Uploads

Use `FormFile` to attach a file from an `io.Reader`, or `FormFileBytes` to attach a file from a byte slice.

### From an io.Reader

```go
httpc := httpr.NewClient()

file, err := os.Open("document.pdf")
if err != nil {
    // Handle error
}
defer file.Close()

resp, err := httpc.Post(
    context.Background(),
    "https://api.example.com/upload",
    httpr.RequestBodyMultipart(
        httpr.FormFile("document", "document.pdf", file),
    ),
)
```

### From Bytes

```go
httpc := httpr.NewClient()

imageData, err := os.ReadFile("photo.png")
if err != nil {
    // Handle error
}

resp, err := httpc.Post(
    context.Background(),
    "https://api.example.com/upload",
    httpr.RequestBodyMultipart(
        httpr.FormFileBytes("avatar", "photo.png", imageData),
    ),
)
```

## Mixed Fields and Files

Fields and files can be freely combined in a single request.

```go
httpc := httpr.NewClient()

file, err := os.Open("photo.jpg")
if err != nil {
    // Handle error
}
defer file.Close()

resp, err := httpc.Post(
    context.Background(),
    "https://api.example.com/profile",
    httpr.RequestBodyMultipart(
        httpr.FormField("username", "moe"),
        httpr.FormField("role", "admin"),
        httpr.FormFile("avatar", "photo.jpg", file),
    ),
)
```

## Custom Multipart Writer

For advanced use cases that need full control over the `multipart.Writer`, use `RequestBodyMultipartFunc`.

```go
httpc := httpr.NewClient()

resp, err := httpc.Post(
    context.Background(),
    "https://api.example.com/custom",
    httpr.RequestBodyMultipartFunc(func(w *multipart.Writer) error {
        if err := w.WriteField("metadata", `{"type":"report"}`); err != nil {
            return err
        }

        part, err := w.CreateFormFile("file", "report.csv")
        if err != nil {
            return err
        }

        _, err = part.Write(csvBytes)
        return err
    }),
)
```

:::note
When using `RequestBodyMultipartFunc`, the writer is automatically closed after your callback returns. Do not call `w.Close()` yourself.
:::
