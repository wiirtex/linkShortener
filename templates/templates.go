package htmlTemplates

import (
	"html/template"

	"github.com/sirupsen/logrus"
)

var tmpls = template.New("")

func GetTmpls() *template.Template {
	return tmpls
}

func init() {
	_, err := tmpls.ParseFiles("templates/index.html",
		"templates/checkLink.html",
		"templates/error.html",
		"templates/notfound.html",
		"templates/success.html")
	if err != nil {
		logrus.Fatal("Error parsing templates: ", err)
	}
}
