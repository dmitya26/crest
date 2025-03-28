package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
)

const workdir = "/home/dmitri/repos/crawl-tester/"

func printStream(stdout string, stderr string) {
	fmt.Fprintln(os.Stdout, "\033[32m"+stdout+"\033[0m")
	fmt.Fprintln(os.Stderr, "\033[31m"+stderr+"\033[0m")
}

func command(args []string, ctx *Context) {
	err := Handle(args, ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func handleHtml(route string, path string) {
	http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("Template Initialized.")
		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("Template Executed.")

	})
}

func HttpTestsiteRun(wg *sync.WaitGroup) http.Server {
	handleHtml("/", "views/index.html")
	handleHtml("/ActuallyExists", "views/ActuallyExists.html")
	handleHtml("/AnotherWorkingSite", "views/AnotherWorkingSite.html")
	handleHtml("/DoesNotExist", "views/DoesNotExist.html")
	http.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		content, err := ioutil.ReadFile("views/robots.txt")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write(content)
	})
	srv := http.Server{Addr: ":8080"}

	wg.Add(1)
	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	wg.Done()

	return srv
}

// Just some basic happy tests, which works for now. I will add more comprehensive tests in the future.
func TestHttpCrawlingFlags(t *testing.T) {
	var wg sync.WaitGroup
	var ctx Context
	var err error

	srv := HttpTestsiteRun(&wg)

	args := []string{"crest", "-tfv", "http://localhost:8080"}
	err = Handle(args, &ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}

	args = []string{"crest", "-tv", "http://localhost:8080"}
	err = Handle(args, &ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}

	srv.Close()
	wg.Wait()
}
