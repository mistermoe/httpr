---
title: Request Inspection
tableOfContents: true
slug: 1.0/inspect
---

Similar to using `curl -v`, it can often be useful to inspect the raw request and response when debugging or understanding the behavior of a server.

the `httpr.Inspect` option can be used to log the request and response to stdout. This option can be set globally when creating a client or per request

### Global

```go {2}
httpc := httpr.NewClient(
  httpr.Inspect(),
)

resp, err := httpc.Get(context.Background(), "https://jsonplaceholder.typicode.com/posts/1")
```

### Per Request

```go {4}
httpc := httpr.NewClient()

resp, err := httpc.Get(context.Background(), "https://jsonplaceholder.typicode.com/posts/1", 
  httpr.Inspect(),
)
```

### Output

inspecting a `GET` request to `https://jsonplaceholder.typicode.com/posts/1` will produce the following output:

```plaintext
Request:
GET /posts/1 HTTP/1.1
Host: jsonplaceholder.typicode.com
User-Agent: Go-http-client/1.1
Accept-Encoding: gzip


Response:
HTTP/2.0 200 OK
Access-Control-Allow-Credentials: true
Age: 1325
Alt-Svc: h3=":443"; ma=86400
Cache-Control: max-age=43200
Cf-Cache-Status: HIT
Cf-Ray: 8cca10deeae216d0-IAH
Content-Type: application/json; charset=utf-8
Date: Thu, 03 Oct 2024 04:07:32 GMT
Etag: W/"124-yiKdLzqO5gfBrJFrcdJ8Yq0LGnU"
Expires: -1
Nel: {"report_to":"heroku-nel","max_age":3600,"success_fraction":0.005,"failure_fraction":0.05,"response_headers":["Via"]}
Pragma: no-cache
Report-To: {"group":"heroku-nel","max_age":3600,"endpoints":[{"url":"https://nel.heroku.com/reports?ts=1727095511&sid=e11707d5-02a7-43ef-b45e-2cf4d2036f7d&s=NGGjjBYeTfBRqq5CLZIxTDJ3FZ8%2F95hmFn9b1UR4xQ4%3D"}]}
Reporting-Endpoints: heroku-nel=https://nel.heroku.com/reports?ts=1727095511&sid=e11707d5-02a7-43ef-b45e-2cf4d2036f7d&s=NGGjjBYeTfBRqq5CLZIxTDJ3FZ8%2F95hmFn9b1UR4xQ4%3D
Server: cloudflare
Vary: Origin, Accept-Encoding
Via: 1.1 vegur
X-Content-Type-Options: nosniff
X-Powered-By: Express
X-Ratelimit-Limit: 1000
X-Ratelimit-Remaining: 999
X-Ratelimit-Reset: 1727095568

{
  "userId": 1,
  "id": 1,
  "title": "sunt aut facere repellat provident occaecati excepturi optio reprehenderit",
  "body": "quia et suscipit\nsuscipit recusandae consequuntur expedita et cum\nreprehenderit molestiae ut ut quas totam\nnostrum rerum est autem sunt rem eveniet architecto"
}
```
