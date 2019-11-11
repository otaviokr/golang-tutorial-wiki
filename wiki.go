/** 2019 - Otavio K. R.
 * This is the result of tutorial: https://golang.org/doc/articles/wiki/
 *
 * The "other tasks" section are also implemented.
 */

package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"os"
	"path/filepath"
)

// Page is the representation of a page from the wiki.
// Published has the same content from Body, but it is HTML safe.
type Page struct {
	Title string
	Body []byte
	Published template.HTML
}

// To avoid issues finding the template files.
var cwd, _ = os.Getwd()

func (p *Page) save() error {
	filename := filepath.Join(cwd, "data", p.Title + ".txt")
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := filepath.Join(cwd, "data", title + ".txt")
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Find notation to links in the Body text.
	r := regexp.MustCompile("\\[([a-zA-Z0-9]+)\\]")

	// Convert the link notation to a valid HTML tag.
	published := r.ReplaceAllFunc(body, func(s []byte) []byte {
		m := string(s[1 : len(s)-1])
		return []byte("<a href=\"/view/" + m + "\">" + m + "</a>")
	})

	return &Page{Title: title, Body: body, Published: template.HTML(published)}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body), Published: template.HTML(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func frontPageHandler(w http.ResponseWriter, r *http.Request, title string) {
	// Just redirect to a specific page (used when accessing the root address).
	http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

var templates = template.Must(template.ParseFiles(
	filepath.Join(cwd, "tmpl", "edit.html"), 
	filepath.Join(cwd, "tmpl", "view.html")))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/((edit|save|view)/([a-zA-Z0-9]+))?$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[3])
	}
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/", makeHandler(frontPageHandler))
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}
