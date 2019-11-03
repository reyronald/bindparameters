package bindparameters

import (
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
)

func bindChiParametersInto(r *http.Request, fn interface{}) (string, string) {
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
	returnValues := Into(r, getURLParam, fn)
	if lenV := len(returnValues); lenV == 0 {
		return "", ""
	} else if lenV == 1 {
		return returnValues[0].Interface().(string), ""
	} else {
		return returnValues[0].Interface().(string), returnValues[1].Interface().(string)
	}
}

type application struct {
	Router *chi.Mux
}

func newApp() *application {
	router := chi.NewRouter()
	router.Use(render.SetContentType(render.ContentTypeJSON))
	return &application{Router: router}
}

func TestURLParameters(t *testing.T) {
	router := newApp().Router

	router.Get("/user/{id}/post/{postId}", func(w http.ResponseWriter, r *http.Request) {
		bindChiParametersInto(r, func(params struct {
			ID     int `json:"id"`
			PostID int `json:"postId"`
		}) {
			render.JSON(w, r, params)
		})
	})

	apitest.New().
		Handler(router).
		Get("/user/1234/post/9876").
		Expect(t).
		Body(`{"id": 1234, "postId": 9876}`).
		Status(http.StatusOK).
		End()
}

func TestQueryStringOfSimpleTypes(t *testing.T) {
	router := newApp().Router

	router.Get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		bindChiParametersInto(r, func(params struct {
			ID         int    `json:"id"`
			FilterInt  int    `json:"filterInt"`
			FilterStr  string `json:"filterStr"`
			FilterBool bool   `json:"filterBool"`
		}) {
			render.JSON(w, r, params)
		})
	})

	// GET /user/1234
	apitest.New().
		Handler(router).
		Get("/user/1234").
		Expect(t).
		Body(`{"id":1234,"filterInt":0,"filterStr":"","filterBool":false}` + "\n").
		Status(http.StatusOK).
		End()

	// GET /user/1234?filterInt=10
	apitest.New().
		Handler(router).
		Get("/user/1234").
		Query("filterInt", "10").
		Expect(t).
		Body(`{"id":1234,"filterInt":10,"filterStr":"","filterBool":false}` + "\n").
		Status(http.StatusOK).
		End()

	// GET /user/1234?filterInt=10&filterStr=hello
	apitest.New().
		Handler(router).
		Get("/user/1234").
		Query("filterInt", "10").
		Query("filterStr", "hello").
		Expect(t).
		Body(`{"id":1234,"filterInt":10,"filterStr":"hello","filterBool":false}` + "\n").
		Status(http.StatusOK).
		End()

	// GET /user/1234?filterInt=20&filterStr=hello&filterBool=true
	apitest.New().
		Handler(router).
		Get("/user/1234").
		Query("filterInt", "20").
		Query("filterStr", "hello").
		Query("filterBool", "true").
		Expect(t).
		Body(`{"id":1234,"filterInt":20,"filterStr":"hello","filterBool":true}` + "\n").
		Status(http.StatusOK).
		End()
}
func TestQueryStringOfSlices(t *testing.T) {
	router := newApp().Router

	router.Get("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		bindChiParametersInto(r, func(params struct {
			ID            int      `json:"id"`
			FilterArrInt  []int    `json:"filterArrInt"`
			FilterArrStr  []string `json:"filterArrStr"`
			FilterArrBool []bool   `json:"filterArrBool"`
		}) {
			render.JSON(w, r, params)
		})
	})

	// GET /user/1234
	apitest.New().
		Handler(router).
		Get("/user/1234").
		Expect(t).
		Body(`{"id":1234,"filterArrInt":[],"filterArrStr":[],"filterArrBool":[]}` + "\n").
		Status(http.StatusOK).
		End()

	// GET /user/1234?filterArrInt=1
	// GET /user/1234?filterArrInt[]=1
	apitest.New().
		Handler(router).
		Get("/user/1234").
		Query("filterArrInt", "1").
		Expect(t).
		Body(`{"id":1234,"filterArrInt":[1],"filterArrStr":[],"filterArrBool":[]}` + "\n").
		Status(http.StatusOK).
		End()

	// GET /user/1234?filterArrInt=1&filterArrInt=2
	// GET /user/1234?filterArrInt=1[]&filterArrInt[]=2
	apitest.New().
		Handler(router).
		Get("/user/1234").
		Query("filterArrInt", "1").
		Query("filterArrInt", "2").
		Expect(t).
		Body(`{"id":1234,"filterArrInt":[1,2],"filterArrStr":[],"filterArrBool":[]}` + "\n").
		Status(http.StatusOK).
		End()

	// GET /user/1234?filterArrInt=1&filterArrInt=2&filterArrStr=one&filterArrStr=two&filterArrBool=true&filterArrBool=false
	// GET /user/1234?filterArrInt[]=1&filterArrInt[]=2&filterArrStr[]=one&filterArrStr[]=two&filterArrBool[]=true&filterArrBool[]=false
	apitest.New().
		Handler(router).
		Get("/user/1234").
		// Ints
		Query("filterArrInt", "1").
		Query("filterArrInt", "2").
		// Strings
		Query("filterArrStr", "one").
		Query("filterArrStr", "two").
		// Bools
		Query("filterArrBool", "true").
		Query("filterArrBool", "false").
		Expect(t).
		Body(`{"id":1234,"filterArrInt":[1,2],"filterArrStr":["one","two"],"filterArrBool":[true,false]}` + "\n").
		Status(http.StatusOK).
		End()
}

func TestRequestBody(t *testing.T) {
	router := newApp().Router

	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	router.Post("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		bindChiParametersInto(r, func(params struct {
			ID int `json:"id"`
		}, user User) {
			response := struct {
				ID   int  `json:"id"`
				User User `json:"user"`
			}{
				ID:   params.ID,
				User: user,
			}
			render.JSON(w, r, response)
		})
	})

	// POST /user/1234
	apitest.New().
		Handler(router).
		Post("/user/1234").
		JSON(`{"name":"Ronald","age":27}`).
		Expect(t).
		Body(`{"id":1234,"user":{"name":"Ronald","age":27}}` + "\n").
		Status(http.StatusOK).
		End()

}

func TestReturnValues(t *testing.T) {
	router := newApp().Router

	router.Get("/user/{id}/post/{postId}", func(w http.ResponseWriter, r *http.Request) {
		s, ss := bindChiParametersInto(r, func(params struct {
			ID     int `json:"id"`
			PostID int `json:"postId"`
		}) (string, string) {
			render.JSON(w, r, params)
			return "hello", "world"
		})

		assert.Equal(t, s, "hello")
		assert.Equal(t, ss, "world")
	})

	apitest.New().
		Handler(router).
		Get("/user/1234/post/9876").
		Expect(t).
		Body(`{"id": 1234, "postId": 9876}`).
		Status(http.StatusOK).
		End()
}
