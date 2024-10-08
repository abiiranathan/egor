package gor_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"text/template"

	"github.com/abiiranathan/gor/gor"
)

func TestRouterServeHTTP(t *testing.T) {
	r := gor.NewRouter()
	r.Get("/test", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "test")
	})
	r.Get("/test2", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "test2")
	})
	r.Get("/test3", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "test3")
	})
	r.Post("/test4", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "test4")
	})
	r.Put("/test5", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "test5")
	})
	r.Delete("/test6", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "test6")
	})
	r.Patch("/test7", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "test7")
	})
	r.Options("/test8", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "test8")
	})
	r.Head("/test9", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "test9")
	})
	r.Connect("/test10", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "test10")
	})
	r.Trace("/test11", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "test11")
	})

	tests := []struct {
		name     string
		method   string
		path     string
		expected string
	}{
		{"GET", "GET", "/test", "test"},
		{"GET", "GET", "/test2", "test2"},
		{"GET", "GET", "/test3", "test3"},
		{"POST", "POST", "/test4", "test4"},
		{"PUT", "PUT", "/test5", "test5"},
		{"DELETE", "DELETE", "/test6", "test6"},
		{"PATCH", "PATCH", "/test7", "test7"},
		{"OPTIONS", "OPTIONS", "/test8", "test8"},
		{"HEAD", "HEAD", "/test9", "test9"},
		{"CONNECT", "CONNECT", "/test10", "test10"},
		{"TRACE", "TRACE", "/test11", "test11"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			r.ServeHTTP(w, req)
			if w.Body.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, w.Body.String())
			}
		})
	}
}

// test 404
func TestRouterNotFound(t *testing.T) {
	r := gor.NewRouter()
	r.Get("/path", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "test")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/notfound", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// Use a derived type. Form processing should still pass.
type Age int

type User struct {
	Name string `form:"name"`
	Age  Age    `form:"age"`
}

// test sending and reading form data
func TestRouterUrlEncodedFormData(t *testing.T) {
	r := gor.NewRouter()
	r.Post("/urlencoded", func(w http.ResponseWriter, req *http.Request) {
		u := User{}
		err := gor.BodyParser(req, &u)
		if err != nil {
			gor.SendString(w, err.Error())
			return
		}
		gor.SendString(w, u.Name)
	})

	form := url.Values{}
	form.Add("name", "Abiira Nathan")
	form.Add("age", "23")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/urlencoded"+"?"+form.Encode(), nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "Abiira Nathan" {
		t.Errorf("expected Abiira Nathan, got %s", w.Body.String())
	}
}

// test sending and reading json data
func TestRouterJSONData(t *testing.T) {
	r := gor.NewRouter()
	r.Post("/json", func(w http.ResponseWriter, req *http.Request) {
		u := User{}
		err := gor.BodyParser(req, &u)
		if err != nil {
			gor.SendString(w, err.Error())
			return
		}
		gor.SendJSON(w, u)
	})

	u := User{
		Name: "Abiira Nathan",
		Age:  23,
	}

	body, err := json.Marshal(u)
	if err != nil {
		t.Error(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/json", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var u2 User
	json.NewDecoder(w.Body).Decode(&u2)

	if !reflect.DeepEqual(u, u2) {
		t.Errorf("expected %v, got %v", u, u2)
	}

}

func TestBodyParserDerivedTypes(t *testing.T) {
	r := gor.NewRouter()
	r.Post("/json", func(w http.ResponseWriter, req *http.Request) {
		u := User{}
		err := gor.BodyParser(req, &u)
		if err != nil {
			gor.SendString(w, err.Error())
			return
		}
		gor.SendJSON(w, u)
	})

	u := User{
		Name: "Abiira Nathan",
		Age:  23,
	}

	body, err := json.Marshal(u)
	if err != nil {
		t.Error(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/json", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var u2 User
	json.NewDecoder(w.Body).Decode(&u2)

	if !reflect.DeepEqual(u, u2) {
		t.Errorf("expected %v, got %v", u, u2)
	}

}

// multipart/form-data
func TestRouterMultipartFormData(t *testing.T) {
	r := gor.NewRouter()
	r.Post("/multipart", func(w http.ResponseWriter, req *http.Request) {
		u := User{}
		err := gor.BodyParser(req, &u)
		if err != nil {
			gor.SendString(w, err.Error())
			return
		}
		gor.SendString(w, u.Name)
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "Abiira Nathan")
	writer.WriteField("age", "23")
	writer.Close()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/multipart", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "Abiira Nathan" {
		t.Errorf("expected Abiira Nathan, got %s", w.Body.String())
	}
}

// multipart/form-data with file
func TestRouterMultipartFormDataWithFile(t *testing.T) {
	r := gor.NewRouter()
	r.Post("/upload", func(w http.ResponseWriter, req *http.Request) {
		req.ParseMultipartForm(req.ContentLength)
		_, fileHeader, err := req.FormFile("file")
		if err != nil {
			gor.SendString(w, err.Error())
			return
		}

		mpf, err := fileHeader.Open()
		if err != nil {
			gor.SendString(w, err.Error())
			return
		}
		defer mpf.Close()

		buf := &bytes.Buffer{}
		_, err = buf.ReadFrom(mpf)
		if err != nil {
			gor.SendString(w, err.Error())
			return
		}
		w.Write(buf.Bytes())
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Error(err)
	}

	_, err = part.Write([]byte("hello world"))
	if err != nil {
		t.Error(err)
	}

	// close writer before creating request
	writer.Close()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	data, err := io.ReadAll(w.Body)
	if err != nil {
		t.Error(err)
	}

	if string(data) != "hello world" {
		t.Errorf("expected hello world, got %s", string(data))
	}
}

type contextType string

const authContextKey contextType = "auth"

// test route middleware
func TestRouterMiddleware(t *testing.T) {
	r := gor.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			context := context.WithValue(req.Context(), authContextKey, "johndoe")
			req = req.WithContext(context)
			next.ServeHTTP(w, req)
		})
	})

	r.Get("/middleware", func(w http.ResponseWriter, req *http.Request) {
		auth, ok := req.Context().Value(authContextKey).(string)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			gor.SendString(w, "no auth")
			return
		}
		gor.SendString(w, auth)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/middleware", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "johndoe" {
		t.Errorf("expected johndoe, got %s", w.Body.String())
	}
}

const msgKey contextType = "message"

// test chaining of middlewares
func TestRouterChainMiddleware(t *testing.T) {
	r := gor.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			gor.SetContextValue(req, msgKey, "first")
			next.ServeHTTP(w, req)
		})
	})

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			message, ok := gor.GetContextValue(req, msgKey).(string)
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				gor.SendString(w, "no message")
				return
			}

			gor.SetContextValue(req, msgKey, message+" second")
			next.ServeHTTP(w, req)
		})
	})

	r.Get("/chain",
		func(w http.ResponseWriter, req *http.Request) {
			message, ok := req.Context().Value(msgKey).(string)
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				gor.SendString(w, "no message")
				return
			}
			gor.SendString(w, message)
		}, func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				message, ok := gor.GetContextValue(req, msgKey).(string)
				if !ok {
					w.WriteHeader(http.StatusInternalServerError)
					gor.SendString(w, "no message")
					return
				}
				gor.SetContextValue(req, msgKey, message+" third")
				h.ServeHTTP(w, req)
			})
		})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/chain", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "first second third" {
		t.Errorf("expected first second third, got %s", w.Body.String())
	}
}

// test render with a base layout
func TestRouterRenderWithBaseLayout(t *testing.T) {
	templ, err := gor.ParseTemplatesRecursive("../cmd/server/templates",
		template.FuncMap{"upper": strings.ToUpper}, ".html")

	if err != nil {
		panic(err)
	}

	r := gor.NewRouter(
		gor.BaseLayout("base.html"),
		gor.ContentBlock("Content"),
		gor.PassContextToViews(true),
		gor.WithTemplates(templ),
	)

	r.Get("/home_page", func(w http.ResponseWriter, req *http.Request) {
		data := gor.Map{
			"Title": "Home Page",
			"Body":  "Welcome to the home page",
		}

		// Router is accessed in context and used for rending. Same as r.Render()
		// but this way you don't need r in scope.
		r.Render(w, req, "home.html", data)

		if err != nil {
			gor.SendError(w, req, err, http.StatusInternalServerError)
		}
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/home_page", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

}

func CopyDir(src, dst string) error {
	// create the destination directory
	err := os.MkdirAll(dst, 0755)
	if err != nil {
		return err
	}

	// get a list of all the files in the source directory
	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// copy each file to the destination directory
	for _, file := range files {
		srcFile := filepath.Join(src, file.Name())
		dstFile := filepath.Join(dst, file.Name())

		// if the file is a directory, copy it recursively
		if file.IsDir() {
			err = CopyDir(srcFile, dstFile)
			if err != nil {
				return err
			}
		} else {
			// copy the file
			input, err := os.ReadFile(srcFile)
			if err != nil {
				return err
			}
			err = os.WriteFile(dstFile, input, 0644)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func TestRouterStatic(t *testing.T) {
	dirname, err := os.MkdirTemp("", "static")
	if err != nil {
		t.Fatalf("could not create temp dir: %v", err)
	}
	defer os.RemoveAll(dirname)

	file := filepath.Join(dirname, "test.txt")
	err = os.WriteFile(file, []byte("hello world"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	r := gor.NewRouter()
	r.Static("/static", dirname)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/static/notfound.txt", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/static/test.txt", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	data, err := io.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "hello world" {
		t.Errorf("expected hello world, got %s", string(data))
	}

}

func TestRouterFile(t *testing.T) {
	// create a temporary directory for the views
	dirname, err := os.MkdirTemp("", "static")
	if err != nil {
		t.Fatalf("could not create temp dir: %v", err)
	}
	defer os.RemoveAll(dirname)

	file := filepath.Join(dirname, "test.txt")
	err = os.WriteFile(file, []byte("hello world"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	r := gor.NewRouter()
	r.File("/static/test.txt", file)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/static/test.txt", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	data, err := io.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "hello world" {
		t.Errorf("expected hello world, got %s", string(data))
	}
}

// Test route groups
func TestRouterGroup(t *testing.T) {
	r := gor.NewRouter()
	admin := r.Group("/admin")

	admin.Get("/home", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "test")
	})

	admin.Get("/users", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "test2")
	})

	// test /admin/test
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/home", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "test" {
		t.Errorf("expected test, got %s", w.Body.String())
	}

	// test /admin/test2
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/admin/users", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "test2" {
		t.Errorf("expected test2, got %s", w.Body.String())
	}
}

// test groups with middleware
func TestRouterGroupMiddleware(t *testing.T) {
	r := gor.NewRouter()
	admin := r.Group("/admin", func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			gor.SetContextValue(req, "admin", "admin middleware")
			next.ServeHTTP(w, req)
		})
	})

	admin.Get("/test", func(w http.ResponseWriter, req *http.Request) {
		admin, ok := gor.GetContextValue(req, "admin").(string)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			gor.SendString(w, "no admin")
			return
		}
		gor.SendString(w, admin)
	})

	// test /admin/test
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "admin middleware" {
		t.Errorf("expected admin middleware, got %s", w.Body.String())
	}
}

// test nested groups
func TestRouterNestedGroup(t *testing.T) {
	r := gor.NewRouter()
	admin := r.Group("/admin")
	users := admin.Group("/users")

	users.Get("/test", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "test")
	})

	// test /admin/users/test
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/users/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "test" {
		t.Errorf("expected test, got %s", w.Body.String())
	}
}

// test gor.Redirect
func TestRouterRedirect(t *testing.T) {
	r := gor.NewRouter()
	r.Get("/redirect1", func(w http.ResponseWriter, req *http.Request) {
		gor.Redirect(w, req, "/redirect2", http.StatusFound)
	})

	r.Get("/redirect2", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "redirect2")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/redirect1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected status 302, got %d", w.Code)
	}

}

// test redirect route
func TestRouterRedirectRoute(t *testing.T) {
	r := gor.NewRouter()
	r.Get("/redirect_route1", func(w http.ResponseWriter, req *http.Request) {
		r.RedirectRoute(w, req, "/redirect_route2", http.StatusFound)
	})

	r.Get("/redirect_route2", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "redirect_route2")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/redirect_route1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected status 302, got %d", w.Code)
	}

}

/*

func Query(req *http.Request, key string, defaults ...string) string {
	v := req.URL.Query().Get(key)
	if v == "" && len(defaults) > 0 {
		return defaults[0]
	}
	return v
}

// paramInt returns the value of the parameter as an integer
func ParamInt(req *http.Request, key string, defaults ...int) int {
	v := req.PathValue(key)
	if v == "" && len(defaults) > 0 {
		return defaults[0]
	}

	vInt, err := strconv.Atoi(v)
	if err != nil {
		if len(defaults) > 0 {
			return defaults[0]
		}
		return 0
	}
	return vInt
}

// queryInt returns the value of the query as an integer
func QueryInt(req *http.Request, key string, defaults ...int) int {
	v := Query(req, key)
	if v == "" && len(defaults) > 0 {
		return defaults[0]
	}

	vInt, err := strconv.Atoi(v)
	if err != nil {
		if len(defaults) > 0 {
			return defaults[0]
		}
		return 0
	}
	return vInt
}

// save file
func SaveFile(fh *multipart.FileHeader, dst string) error {
	src, err := fh.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}
*/

// test Query
func TestRouterQuery(t *testing.T) {
	r := gor.NewRouter()
	r.Get("/query", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, gor.Query(req, "name", "default"))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/query?name=abiira", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "abiira" {
		t.Errorf("expected abiira, got %s", w.Body.String())
	}
}

// test QueryInt
func TestRouterQueryInt(t *testing.T) {
	r := gor.NewRouter()
	r.Get("/queryint", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, strconv.Itoa(gor.QueryInt(req, "age", 0)))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/queryint?age=23", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "23" {
		t.Errorf("expected 23, got %s", w.Body.String())
	}
}

// test ParamInt
func TestRouterParamInt(t *testing.T) {
	r := gor.NewRouter()
	r.Get("/paramint/{age}", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, req.PathValue("age"))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/paramint/30", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "30" {
		t.Errorf("expected 30, got %s", w.Body.String())
	}
}

// Write a benchmark test for the router
func BenchmarkRouter(b *testing.B) {
	r := gor.NewRouter()
	r.Get("/benchmark", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "Hello World!")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/benchmark", nil)

	for i := 0; i < b.N; i++ {
		r.ServeHTTP(w, req)
	}
}

// bench mark full request/response cycle
func BenchmarkRouterFullCycle(b *testing.B) {
	r := gor.NewRouter()
	r.Get("/benchmark-cycle", func(w http.ResponseWriter, req *http.Request) {
		gor.SendString(w, "Hello World!")
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	for i := 0; i < b.N; i++ {
		res, err := http.Get(ts.URL + "/benchmark-cycle")
		if err != nil {
			b.Fatal(err)
		}
		if res.StatusCode != http.StatusOK {
			b.Fatalf("expected status 200, got %d", res.StatusCode)
		}
	}
}

func TestRouterExecuteTemplate(t *testing.T) {
	templ, err := gor.ParseTemplatesRecursive("../cmd/server/templates",
		template.FuncMap{"upper": strings.ToUpper}, ".html")

	if err != nil {
		panic(err)
	}

	r := gor.NewRouter(gor.WithTemplates(templ))

	r.Get("/template", func(w http.ResponseWriter, req *http.Request) {
		data := gor.Map{
			"Title": "Template",
			"Body":  "Welcome to the template page",
		}

		err := r.ExecuteTemplate(w, "home.html", data)
		if err != nil {
			gor.SendError(w, req, err, http.StatusInternalServerError)
		}

		// Test lookup template
		templ, err = gor.LookupTemplate(req, "home.html")
		if err != nil {
			t.Errorf("expected to find home.html template")
		}

		out := new(bytes.Buffer)
		err = templ.Execute(out, map[string]any{
			"Title": "Template",
			"Body":  "Named Template",
		})

		if err != nil {
			t.Errorf("execute template failed")
		}

		if !strings.Contains(out.String(), "Named Template") {
			t.Errorf("expected 'Named Template' in templated page, got %s", out.String())
		}
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/template", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// check body
	if !strings.Contains(w.Body.String(), "Welcome to the template page") {
		t.Errorf("expected Welcome to the template page, got %s", w.Body.String())
	}

}

func TestRouterExecute(t *testing.T) {
	templ, err := gor.ParseTemplatesRecursive("../cmd/server/templates",
		template.FuncMap{"upper": strings.ToUpper}, ".html")

	if err != nil {
		panic(err)
	}

	r := gor.NewRouter(gor.WithTemplates(templ))

	r.Get("/template", func(w http.ResponseWriter, req *http.Request) {
		data := gor.Map{
			"Title": "Template",
			"Body":  "Welcome to the template page",
		}

		err := gor.Execute(w, req, "home.html", data)
		if err != nil {
			gor.SendError(w, req, err, http.StatusInternalServerError)
		}

	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/template", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// check body
	if !strings.Contains(w.Body.String(), "Welcome to the template page") {
		t.Errorf("expected Welcome to the template page, got %s", w.Body.String())
	}

}

func TestRouterFileFS(t *testing.T) {
	dirname, err := os.MkdirTemp("", "assets")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirname)

	file := filepath.Join(dirname, "test.txt")
	err = os.WriteFile(file, []byte("hello world"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	r := gor.NewRouter()
	r.FileFS(http.Dir(dirname), "/static", "test.txt")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/static", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	data, err := io.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "hello world" {
		t.Errorf("expected hello world, got %s", string(data))
	}
}

func TestRouterFaviconFS(t *testing.T) {
	dirname, err := os.MkdirTemp("", "assets")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirname)

	file := filepath.Join(dirname, "favicon.ico")
	err = os.WriteFile(file, []byte("hello world"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	r := gor.NewRouter()
	r.FaviconFS(http.Dir(dirname), "favicon.ico")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/favicon.ico", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	data, err := io.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "hello world" {
		t.Errorf("expected hello world, got %s", string(data))
	}
}
