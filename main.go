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
	p, err := readFile("pages.txt", false)
	_ = err
	r, err := readFile("resources.txt", false)
	if err != nil {

	}
	for x := range strings.Lines(p) {
		pages = append(pages, strings.ReplaceAll(x, "\n", ""))
	}
	for x := range strings.Lines(r) {
		files = append(files, strings.ReplaceAll(x, "\n", ""))
	}
}

func renderFile(name string, w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	w.Header().Add("Cache-Content", "no-cache")
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
	var err error
	if r.RequestURI == "/" {
		x, err = readFile("index.html", true)
		if err != nil {
			writePage("", w)
			return
		}
	} else {
		if strings.Contains(r.RequestURI, "favicon") {
			return
		}
		// x, err = readFile(r.RequestURI[1:], true)
		x, err = readFile(page, true)
		if err != nil {
			writePage("", w)
			return
		}
	}
	_ = x
	h, b, f := retreiveHBF(page)
	header := fmt.Sprintf(h, page)
	template, err := readFile("template.html", true)
	if err != nil {
		log.Println("Template:", err)
	}
	
	s := fmt.Sprintf(template, header, b, f)
	writePage(s, w)
}

func retreiveHBF(name string) (string, string, string) {
	head, err := readFile("head.html", true)
	_ = err
	body, err := readFile(name, true)
	_ = err
	footer, err := readFile("footer.html", true)
	if err != nil {
		return "", "", ""
	}
	return head, body, footer
}

func readFile(s string, useFS bool) (string, error) {
	if useFS {
		file, err := fs.Open(s)
		if err != nil {
			log.Println(err)
			return "", err
		}
		b, err := io.ReadAll(file)
		if err != nil {
			log.Println(err)
			return "", err
		}
		return string(b), nil
	} else {
		file, err := os.Open(s)
		if err != nil {
			log.Println(err)
			return "", err
		}
		b, err := io.ReadAll(file)
		if err != nil {
			log.Println(err)
			return "", err
		}
		return string(b), nil
	}
}

func writePage(s string, w http.ResponseWriter) {
	w.Write([]byte(s))
}

func logRequest(r *http.Request) {
	log.Println(r.RemoteAddr+": ", r.RequestURI)
}
