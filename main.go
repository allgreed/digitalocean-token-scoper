// heavily based on https://github.com/davidfstr/nanoproxy
package main

import (
    "fmt"
    "io"
    "net/http"
    "time"
)


var auth = map[string]string {
    "aaaa": "xxxd",
}


func main() {
    // TODO: get envs here and pass them down?
    // TODO: get the real token from secretpath

    // TODO: populate auth from some kind of config? -> map tokens from secret files based on username to permissions

    handler := http.DefaultServeMux
    handler.HandleFunc("/", handleFunc)
    s := &http.Server{
        // TODO: parametrize serving port
        Addr:           ":8080",
        Handler:        handler,
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   10 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }
    
    s.ListenAndServe()
    // TODO: move it to the bottom
}


func handleFunc(w http.ResponseWriter, r *http.Request) {
    // TODO: log incoming request properly
    //fmt.Printf("--> %v %v\n", r.Method, r.URL)
        // everything except the token
        // give each request some id that'll stick through the logs


    _token := r.Header["Authorization"]
    if len(_token) != 1 {
        http.Error(w, "Please provide a token in the Authorization header", 401)
        // TODO: log - missing or malformed token
        return
    }
    token := _token[0]

    permissions, ok := auth[token]
    if !ok {
        http.Error(w, "Please provide a valid token in the Authorization header", 401)
        // TODO: log - invalid token
        return
    }
    // TODO: log - authenticated as ...
    

    // TODO: make'ify companion scripts for env-dev
        // docker run -p 5678:5678 hashicorp/http-echo -text="hello world"
        // name, demon, for easeri killing
    // TODO: make'ify terminal commends for interacting
        // curl localhost:8080 -H "Authorization: aaaa"
    // TODO: come up with actuall permission format
    fmt.Println("\v", permissions)
    // nah just make classes :D and check each of them if the apply! based on path and method!
    // TODO: do the authorization check
    //r.URL.Path
    //r.Method
    // TODO: test
    // if fails return 403 with info that request path / method is not allowed
    // TODO: log - unauthorized action attempt

    hh := http.Header{}
    for k,v := range r.Header {
      hh[k] = v
    }

    if _, ok := hh["Authorization"]; ok {
        // TODO: append the real token -> when I've got the token
        hh["Authorization"] = []string{"BLE"}
    }


    // TODO paremtrize Scheme default(https) and host-port default("api.digitalocean")
    r.URL.Host = "localhost:5678"
    r.URL.Scheme = "http"

    proxied_request := http.Request{
        Method: r.Method,
        URL: r.URL,
        Header: hh,
        Body: r.Body,
        ContentLength: r.ContentLength,
        Close: r.Close,
    }
    resp, err := http.DefaultTransport.RoundTrip(&proxied_request)
    if err != nil {
        // TODO: relay more info ?
        http.Error(w, "Could not reach origin server", 500)
        // TODO: log - request failed
        return
    }
    defer resp.Body.Close()
    

    respH := w.Header()
    for hk, hv := range resp.Header {
        respH[hk] = hv
    }
    w.WriteHeader(resp.StatusCode)
    if resp.ContentLength > 0 {
        // ignore I/O errors, since there's nothing we can do
        io.CopyN(w, resp.Body, resp.ContentLength)
    } else if (resp.Close) {
        for {
            if _, err := io.Copy(w, resp.Body); err != nil {
                break
            }
        }
    }
    // TODO: log - request completed succesfully
    // TODO: update the README that it's working
    // TODO: build with nix
    // TODO: create minimal container
    // TODO: setup CI

    // TODO: add metrics
    // TODO: add proper usage, etc to the README
}
