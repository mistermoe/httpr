a try-hard general purpose http client for golang. hipster docs [here](https://mistermoe.github.io/httpr/)

![Go Badge](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=fff&style=flat) [![Go Report Card](https://goreportcard.com/badge/github.com/mistermoe/httpr)](https://goreportcard.com/report/github.com/mistermoe/httpr) [![integrity](https://github.com/mistermoe/httpr/actions/workflows/integrity.yml/badge.svg)](https://github.com/mistermoe/httpr/actions/workflows/integrity.yml)

## Table of Contents <!-- omit in toc -->
- [Rationale](#rationale)
- [Getting Started](#getting-started)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Output](#output)
- [Features](#features)


## Rationale
why does this library exist when [all these](https://awesome-go.com/http-clients/) already do? Why not just use the standard library?

I'm in an environment that requires myself and a team to integrate with several money movement vendors to move large amounts of money on a daily basis. These vendors vary from omega hipster startups to evilcorp banks. The APIs surfaced by these vendors vary significantly in many ways e.g. auth, response formats, error handling, etc. 

After trying all of the libraries that provide full featured http clients, we weren't able to cover the spectrum of requirements we have. They all fell a bit short when it came to how interceptors work or how error response bodies are handled.

`httpr` is something that works for me. I'm sharing it in case it works for you too. It comes with a few features that I believe help increase the likelihood of me being a responsible adult: observability

## Getting Started

### Installation

```bash
go get github.com/mistermoe/httpr
```

### Usage

the following is an example of a handful of features that httpr provides. This example sends a POST request to `https://reqres.in/api/register` with a request body and expects a response body of type `RegisterResponse` or `RegisterErrorResponse` in case of an error response.

```go
package main

import (
  "context"
  "fmt"
  "log"
  "net/http"

  "github.com/mistermoe/httpr"
)

type RegisterRequest struct {
  Email    string `json:"email"`
  Password string `json:"password"`
}

type RegisterResponse struct {
  ID    int    `json:"id"`
  Token string `json:"token"`
}

type RegisterErrorResponse struct {
  Error string `json:"error"`
}

func main() {
  client := httpr.NewClient(
    httpr.BaseURL("https://reqres.in"),
    httpr.Inspect(),
  )

  reqBody := RegisterRequest{
    Email: "eve.holt@reqres.in",
    // Email:    "moegrammer@hehe.gov", // uncomment for 400 response
    Password: "wowsosecret",
  }

  var respBody RegisterResponse
  var errBody RegisterErrorResponse

  resp, err := client.Post(
    context.Background(),
    "/api/register",
    httpr.RequestBodyJSON(reqBody),
    httpr.ResponseBodyJSON(&respBody, &errBody),
  )

  if err != nil {
    log.Fatalf("Request failed: %v", err)
  }

  if resp.StatusCode == http.StatusBadRequest {
    fmt.Printf("(%v): %v\n", resp.StatusCode, errBody.Error)
  } else {
    fmt.Printf("registration successful: %v\n", respBody.ID)
  }
}
```

> [!TIP]
> `httpr.Inspect()` is an interceptor that logs the request and response. It's useful for debugging.


#### Output
```
Request:
POST /api/register HTTP/1.1
Host: reqres.in
User-Agent: Go-http-client/1.1
Content-Length: 55
Content-Type: application/json
Accept-Encoding: gzip

{"email":"eve.holt@reqres.in","password":"wowsosecret"}
Response:
HTTP/2.0 200 OK
Content-Length: 36
Access-Control-Allow-Origin: *
Cf-Cache-Status: DYNAMIC
Cf-Ray: 8d2d1d1dbef56bae-DFW
Content-Type: application/json; charset=utf-8
Date: Tue, 15 Oct 2024 04:37:25 GMT
Etag: W/"24-4iP0za1geN2he+ohu8F0FhCjLks"
Nel: {"report_to":"heroku-nel","max_age":3600,"success_fraction":0.005,"failure_fraction":0.05,"response_headers":["Via"]}
Report-To: {"group":"heroku-nel","max_age":3600,"endpoints":[{"url":"https://nel.heroku.com/reports?ts=1728967044&sid=c4c9725f-1ab0-44d8-820f-430df2718e11&s=da4UajPHCv9cP90lDWJTH0yPoeHNweUdOPgmJcavq8s%3D"}]}
Reporting-Endpoints: heroku-nel=https://nel.heroku.com/reports?ts=1728967044&sid=c4c9725f-1ab0-44d8-820f-430df2718e11&s=da4UajPHCv9cP90lDWJTH0yPoeHNweUdOPgmJcavq8s%3D
Server: cloudflare
Via: 1.1 vegur
X-Powered-By: Express

{"id":4,"token":"QpwL5tke4Pnpja7X4"}
registration successful: 4
```

## Features
* BaseURL configuration - set base url for all requests
* Setting custom headers both globally and per request
* Setting custom query params
* Supplying strongly typed request bodies 
* Unmarshalling response bodies into strong types (for success and error responses)
* Interceptor support for request/response modification and inspection
* Built-in request/response inspector for debugging
* Opt-in OLTP Instrumentation with metrics and traces for observability