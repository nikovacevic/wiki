// Wiki serves content for viewing and modifying data via HTTP.

package wiki

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
}

type Directory struct {
	Pages []Page
}

const (
	rootDir = "../"
	tempDir = rootDir + "assets/templates/"
	dataDir = rootDir + "data/"
)

// templates are html templates that are required before serving content
var templates = template.Must(template.ParseFiles(tempDir+"index.html", tempDir+"edit.html", tempDir+"view.html"))

// validPath defines the set of valid URL paths
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// save writes a Page to a file. The file name is the title.
func (p *Page) save() error {
	filename := dataDir + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

// loadPage reads a Page from a file and returns a pointer to a Page literal
// constructed with the values read from file.
func loadPage(title string) (*Page, error) {
	filename := dataDir + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

//
func loadDirectory() (*Directory, error) {
	d := Directory{}
	files, err := ioutil.ReadDir(rootDir + "data/")
	if err != nil {
		return nil, err
	}
	if len(files) <= 0 {
		return nil, nil
	}
	// fmt.Println("./data")
	for _, f := range files {
		if f.Name() == "README.md" {
			continue
		}
		t := f.Name()[:len(f.Name())-len(".txt")]
		// fmt.Println("|-" + t + ": " + f.Name())
		d.Pages = append(d.Pages, Page{Title: t})
	}
	return &d, nil
}

// renderTemplate renders an HTML template, identified by temp
func renderTemplate(w http.ResponseWriter, t string, p *Page) {
	err := templates.ExecuteTemplate(w, t+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// IndexHandler serves a directory listing of Pages
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	d, err := loadDirectory()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if d == nil {
		http.Redirect(w, r, "/edit/page", http.StatusFound)
	}
	err = templates.ExecuteTemplate(w, "index.html", d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// viewHandler serves an HTML document built from a Page, which is loaded
// by title, which is parsed from the request URL
func ViewHandler(w http.ResponseWriter, r *http.Request, t string) {
	p, err := loadPage(t)
	if err != nil {
		http.Redirect(w, r, "/edit/"+t, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

// editHandler serves an HTML form for editing a Page, which is identified
// by title, which is parsed from the request URL
func EditHandler(w http.ResponseWriter, r *http.Request, t string) {
	p, err := loadPage(t)
	if err != nil {
		p = &Page{Title: t}
	}
	renderTemplate(w, "edit", p)
}

// saveHandler creates a Page type from Request data and saves it to
// persistent storage under the title.txt filename convention
func SaveHandler(w http.ResponseWriter, r *http.Request, t string) {
	body := r.FormValue("body")
	p := &Page{Title: t, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/view/"+t, http.StatusFound)
}

// makeHandler wraps an http handler, validates the title and request URL,
// then generates a new http handler
func MakeHandler(f func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := validPath.FindStringSubmatch(r.URL.Path)
		if t == nil {
			http.NotFound(w, r)
			return
		}
		f(w, r, t[2])
	}
}
