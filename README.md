# bindparameters

Exposes a small utility function that will automatically bind URL parameters, query string parameters and/or body JSON payloads from a HTTP request into your own types without any need for manual decoding, marshalling or lookups, through a user-provided callback.

```
go get -u github.com/reyronald/bindparameters
```

## Documentation

### `Into`

```go
func Into(
	r *http.Request,
	getURLParam func(key string) string,
	fn interface{},
)
```

`Into` will automatically bind or map parameters from the HTTP request `r` into the arguments of `fn`. `fn` must be a function with either one or two arguments. The first argument should be a struct with fields that map to URL and query string parameters. The second argument is optional and will be used to bind/map the JSON payload of the request into it. If your endpoint doesn't have any URL or query string parameters (or you don't need to access them in your handler), you still need to provide the first argument to the function, but in that case you can pass `nil`.

Example:

```go
type Post struct {
    ID    int    `json:"int,string"`
    Title string `json:"title"`
    // ... others
}

var getURLParam func(key string) string

router.Put("/user/{id}/post/{postId}", func(w http.ResponseWriter, r *http.Request) {
    bindparameters.Into(r, getURLParam, func(params struct {
        ID     int
        PostID int
    }, post Post) {
        // Here the `params.ID`, `params.PostID` and `post`
        // variables will be automatically populated
    })
})
```

Since Go's native `http` module does not have an API to extract URL parameters you'll have to provide your own user-defined `getURLParam` callback, and its implementation will vary depending on the routing library you are using. You can then simplify the `Into` call if you wrap it around a function. Here's an example of how that could look like using [chi](https://github.com/go-chi/chi):

```go
func bindChiParametersInto(r *http.Request, fn interface{}) {
	getURLParam := func(key string) string {
		if rctx := chi.RouteContext(r.Context()); rctx != nil {
			for k := len(rctx.URLParams.Keys) - 1; k >= 0; k-- {
				if strings.ToLower(rctx.URLParams.Keys[k]) == strings.ToLower(key) {
					return rctx.URLParams.Values[k]
				}
			}
		}

		return ""
	}
	bindparameters.Into(r, getURLParam, fn)
}
```

> NOTE: Check out [./\_examples/main.go](./_examples/main.go) and [./bindparameters_test.go](./bindparameters_test.go) for more usage examples.
