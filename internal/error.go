package errors

import (
	"html/template"
	"net/http"
)

type ErrorData struct {
	Code    int
	Title   string
	Message string
}

var ErrorTmpl *template.Template

func Init(path string) {
	var err error
	ErrorTmpl, err = template.ParseFiles(path)
	if err != nil {
		panic("Failed to load error template: " + err.Error())
	}
}

func RenderError(w http.ResponseWriter, code int, title, message string) {
	w.WriteHeader(code)
	errData := ErrorData{
		Code:    code,
		Title:   title,
		Message: message,
	}
	if err := ErrorTmpl.Execute(w, errData); err != nil {
		http.Error(w, "Error rendering error page", http.StatusInternalServerError)
	}
}
