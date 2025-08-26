package main

import (
	"encoding/json"
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
	jpages      = new(Sitemap)
	files       []string
	blocked     = []string{
		".well-known",
	}
)

func main() {
	setOutput()
	readPages()
	log.Println("Starting...")
	initHandlers()
	initServer()
}

func setOutput() {
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
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
	getLinks()
	for _, v := range jpages.Pages {
		http.HandleFunc(v, func(w http.ResponseWriter, r *http.Request) {
			renderPage(v, w, r)
		})
	}
}

func getLinks() {
	// p, err := readFile("pages.txt", false)
	// _ = err
	r, err := readFile("resources.txt", false)
	if err != nil {

	}
	// for x := range strings.Lines(p) {
	// 	pages = append(pages, strings.ReplaceAll(x, "\n", ""))
	// }
	for x := range strings.Lines(r) {
		files = append(files, strings.ReplaceAll(x, "\n", ""))
	}
}

func renderIndex(w http.ResponseWriter, r *http.Request) {
	renderTemplate("index", w, r)
}

func renderPage(name string, w http.ResponseWriter, r *http.Request) {
	for _, x := range blocked {
		if strings.Contains(r.RequestURI, x) {
			w.WriteHeader(404)
			return
		}
	}
	logRequest(name, r)
	w.Header().Add("Cache-Control", "no-cache")
	// fmt.Printf("Name: %s | URI: %s\n", name, r.RequestURI)
	if r.RequestURI != name { //serve a file
		http.ServeFileFS(w, r, fs, r.RequestURI)
	} else { //serve webpage
		n := name[1:] //
		if n == "" {  //index page
			renderIndex(w, r)
		} else {
			renderTemplate(n, w, r)
		}
	}
}

func renderTemplate(page string, w http.ResponseWriter, r *http.Request) {
	s := formatPage(page)
	writePage(s, w)
}

func formatPage(page string) string {
	h, b, f := retreiveHBF(page)
	header := fmt.Sprintf(h, page)
	template, err := readFile("templates/template.html", true)
	if err != nil {
		log.Println("Template:", err)
	}
	s := fmt.Sprintf(template, header, b, f)
	return s
}

func retreiveHBF(name string) (string, string, string) {
	head, err := readFile("templates/head.html", true)
	_ = err
	body, err := readFile("pages/"+name+".html", true)
	_ = err
	footer, err := readFile("templates/footer.html", true)
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

func logRequest(message string, r *http.Request) {
	log.Printf("%s\n", r.RemoteAddr+":"+r.UserAgent()+" | "+message+" | "+r.RequestURI)
}

func readPages() {
	file, err := os.Open("pages.json")
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(jpages)
	if err != nil {
		log.Println(err)
	}
}

func writePages() {
	file, err := os.OpenFile("pages.json", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(jpages)
	if err != nil {
		log.Println(err)
	}
}

type Sitemap struct {
	Pages []string `json:"pages"`
}

type Webpage struct {
	Name        string
	NestedPages bool
}
