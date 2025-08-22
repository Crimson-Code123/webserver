package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	useTLS      = false
	logFileName = "site.log"
	logger      *os.File
	fsdir       = "contents"
	fs          = os.DirFS(fsdir + "/")
	pages       = []string{}
	files       = []string{}
)

func main() {
	setOutput()
	fmt.Println("")
	log.Println("Starting...")
	initHandlers()
	initServer()
}

func setOutput() {
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Println(err)
	}
	logger = file
	// logger.SetWriteDeadline()
	log.SetOutput(logger)
}

func initServer() {
	var err error
	if useTLS {
		err = http.ListenAndServeTLS("0.0.0.0:8080", "cert.pem", "key.pem", nil)
	} else {
		err = http.ListenAndServe("0.0.0.0:8080", nil)
	}
	if err != nil {
		log.Panic(err)
	}
}

func initHandlers() {
	//create file list and display
	getLinks()
	for _, v := range pages {
		http.HandleFunc(v, func(w http.ResponseWriter, r *http.Request) {
			renderPage(v, w, r)
		})
	}
	for _, v := range files {
		http.HandleFunc(v, func(w http.ResponseWriter, r *http.Request) {
			renderFile(v, w, r)
		})
	}
}

func getLinks() {
	p := readFile("pages.txt", false)
	r := readFile("resources.txt", false)
	for x := range strings.Lines(p) {
		pages = append(pages, strings.ReplaceAll(x, "\n", ""))
	}
	for x := range strings.Lines(r) {
		files = append(files, strings.ReplaceAll(x, "\n", ""))
	}
}

func renderFile(name string, w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	http.ServeFileFS(w, r, fs, name)
}

func renderIndex(w http.ResponseWriter, r *http.Request) {
	renderTemplate("index.html", w, r)
}

func renderPage(name string, w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	if len(name[1:]) == 0 { //index page
		renderIndex(w, r)
	} else {
		renderTemplate(name[1:]+".html", w, r)
	}
}

func renderTemplate(page string, w http.ResponseWriter, r *http.Request) {
	var x string
	if r.RequestURI == "/" {
		x = readFile("index.html", true)
	} else {
		if strings.Contains(r.RequestURI, "favicon") {
			return
		}
		x = readFile(r.RequestURI, true)
	}
	h, b, f := retreiveHBF()
	s := fmt.Sprintf(x, fmt.Sprintf(h, page), b, f)
	writePage(s, w)
}

func retreiveHBF() (string, string, string) {
	head := readFile("head.html", true)
	body := readFile("body.html", true)
	footer := readFile("footer.html", true)
	return head, body, footer
}

func readFile(s string, useFS bool) string {
	// fmt.Println(s)
	if useFS {
		file, err := fs.Open(s)
		if err != nil {
			log.Println(err)
		}
		b, err := io.ReadAll(file)
		if err != nil {
			log.Println(err)
		}
		return string(b)
	} else {
		file, err := os.Open(s)
		if err != nil {
			log.Println(err)
		}
		b, err := io.ReadAll(file)
		if err != nil {
			log.Println(err)
		}
		return string(b)
	}
}

func writePage(s string, w http.ResponseWriter) {
	w.Write([]byte(s))
}

func logRequest(r *http.Request) {
	log.Println(r.RemoteAddr+": ", r.RequestURI)
}
