package main

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/reyronald/bindparameters"
)

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

type user struct {
	Name string `json:"name"`
	Age  int    `json:"age,string,omitempty"`
}

func main() {
	router := chi.NewRouter()
	router.Use(render.SetContentType(render.ContentTypeJSON))

	// http --ignore-stdin GET :7000/user/1/post/1000
	router.Get("/user/{id}/post/{postId}", func(w http.ResponseWriter, r *http.Request) {
		bindChiParametersInto(r, func(params struct {
			ID     int `json:"id"`
			PostID int `json:"postId"`
		}) {
			render.JSON(w, r, params)
		})
	})

	// http --ignore-stdin GET ":7000/query-strings-simple/1?filterInt=25&filterStr=hello&filterBool=true"
	router.Get("/query-strings-simple/{id}", func(w http.ResponseWriter, r *http.Request) {
		bindChiParametersInto(r, func(params struct {
			ID         int    `json:"id"`
			FilterInt  int    `json:"filterInt"`
			FilterStr  string `json:"filterStr"`
			FilterBool bool   `json:"filterBool"`
		}) {
			render.JSON(w, r, params)
		})
	})

	// http --ignore-stdin GET ":7000/query-strings-arrays/1?filterArrInt[]=1&filterArrInt[]=2&filterArrStr[]=hello"
	router.Get("/query-strings-arrays/{id}", func(w http.ResponseWriter, r *http.Request) {
		bindChiParametersInto(r, func(params struct {
			ID            int      `json:"id"`
			FilterArrInt  []int    `json:"filterArrInt"`
			FilterArrStr  []string `json:"filterArrStr"`
			FilterArrBool []bool   `json:"filterArrBool"`
		}) {
			render.JSON(w, r, params)
		})
	})

	// http --ignore-stdin POST :7000/user/1 name=Ronald age=27
	router.Post("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		bindChiParametersInto(r, func(params struct {
			ID int `json:"id"`
		}, u user) {
			response := struct {
				ID   int  `json:"id"`
				User user `json:"user"`
			}{
				ID:   params.ID,
				User: u,
			}
			render.JSON(w, r, response)
		})
	})

	http.ListenAndServe(":7000", router)
}
