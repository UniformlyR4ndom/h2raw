package main

import (
    "crypto/rand"
    "fmt"
    "log"
    "strings"

    "github.com/UniformlyR4ndom/h2raw/h2raw"
)


func main() {
    endpoint, authority := "example.com:443", "example.com"
    // endpoint, authority := "127.0.0.1:8000", "localhost:8000" // local http server

    exampleSimpleGet(endpoint, authority)
    exampleFatGet(endpoint, authority)
    exampleSimplePost(endpoint, authority)
    examplePostLyingContentLength(endpoint, authority)
    exampleJunk(endpoint, authority)
}


// Send a simple HTTP/2 GET request with a custom header and a host header
// Care should be taken with uppercase headers since the 
// receiving server might not like those!
func exampleSimpleGet(endpoint, authority string) {
    client, err := h2raw.NewTLSInsecureClient(endpoint)
    if checkError(err) {
        return
    }

    h2conn, err := client.NewConnection(endpoint)
    if checkError(err) {
        return
    }

    req := h2raw.NewRequest("GET", "https", authority, "/example/simpleget")
    req.AddHeader("host", authority)
    req.AddHeader("x-demo-header", "just some header value")

    resp, err := client.MakeRequest(h2conn, req, 1)
    if checkError(err) {
        return
    }

    log.Printf("Response:\n%v\n\n", respToString(resp))
}


// Send an HTTP/2 GET request with a body
func exampleFatGet(endpoint, authority string) {
    client, err := h2raw.NewTLSInsecureClient(endpoint)
    if checkError(err) {
        return
    }

    h2conn, err := client.NewConnection(endpoint)
    if checkError(err) {
        return
    }

    req := h2raw.NewRequest("GET", "https", authority, "/example/fatget")
    body := []byte("This GET request has a body!")
    req.Body = &body

    resp, err := client.MakeRequest(h2conn, req, 1)
    if checkError(err) {
        return
    }

    log.Printf("Response:\n%v\n\n", respToString(resp))
}


// Send an HTTP/2 POST request with a binary body
func exampleSimplePost(endpoint, authority string) {
    // build body
    pre := []byte("First some text. Next binary data follows.")
    post := []byte("And finally some text again")
    binaryLength := 100
    body := make([]byte, 0, len(pre) + binaryLength + len(post))

    randData := make([]byte, binaryLength)
    rand.Read(randData)
    body = append(body, pre...)
    body = append(body, randData...)
    body = append(body, post...)

    client, err := h2raw.NewTLSInsecureClient(endpoint)
    if checkError(err) {
        return
    }

    h2conn, err := client.NewConnection(endpoint)
    if checkError(err) {
        return
    }

    req := h2raw.NewRequest("POST", "https", authority, "/example/post")
    req.Body = &body

    resp, err := client.MakeRequest(h2conn, req, 1)
    if checkError(err) {
        return
    }

    log.Printf("Response:\n%v\n\n", respToString(resp))
}


// Send an HTTP/2 POST request where the content-length
// header does not match the actual length of the body
func examplePostLyingContentLength(endpoint, authority string) {
    client, err := h2raw.NewTLSInsecureClient(endpoint)
    if checkError(err) {
        return
    }

    h2conn, err := client.NewConnection(endpoint)
    if checkError(err) {
        return
    }

    req := h2raw.NewRequest("POST", "https", authority, "/example/post")
    req.AddHeader("content-length", "2")
    body := []byte("Some body longer than 2 bytes!")
    req.Body = &body

    resp, err := client.MakeRequest(h2conn, req, 1)
    if checkError(err) {
        return
    }

    log.Printf("Response:\n%v\n\n", respToString(resp))
}


// Send an HTTP/2 request with a junk method
func exampleJunk(endpoint, authority string) {
    client, err := h2raw.NewTLSInsecureClient(endpoint)
    if checkError(err) {
        return
    }

    h2conn, err := client.NewConnection(endpoint)
    if checkError(err) {
        return
    }

    req := h2raw.NewRequest("JUNK method 1234", "https", authority, "/example/simpleget")
    req.AddHeader("x-demo-header", "just some header value")

    resp, err := client.MakeRequest(h2conn, req, 1)
    if checkError(err) {
        return
    }

    log.Printf("Response:\n%v\n\n", respToString(resp))
}


func respToString(resp *h2raw.Response) string {
    sb := strings.Builder{}

    for k, values := range(resp.Headers) {
        for _, v := range(values) {
            sb.WriteString(fmt.Sprintf("%v: %v\n", k, v))
        }
    }

    if resp.HasBody() {
        sb.WriteString(fmt.Sprintf("\n%v", string(*resp.Body)))
    }

    return sb.String()
}


func checkFatal(err error) {
    if err != nil {
        log.Fatalf("Error: %v\n", err)
    }
}


func checkError(err error) bool {
    if err != nil {
        log.Printf("Error: %v\n", err)
        return true
    }

    return false
}
