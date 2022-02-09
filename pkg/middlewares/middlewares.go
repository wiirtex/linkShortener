package middlewares

import (
	"context"
	b64 "encoding/base64"
	"linkShortener/pkg/helpers"
	htmlTemplates "linkShortener/templates"
	"net/http"
	"net/url"
	"unsafe"

	"github.com/sirupsen/logrus"
)

var requestId = 0

func newRequest(logEntry *logrus.Entry) *logrus.Entry {
	requestId += 1
	return logEntry.WithField("rqId", requestId-1)
}

func checkLink(link string) (isLink bool, err error) {

	_, err = url.ParseRequestURI(link)
	if err != nil {
		return false, err
	}

	u, err := url.Parse(link)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false, err
	}

	return true, nil
}

func ManageLongLink(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logrus.New()
		logEntry := logrus.NewEntry(logger)
		logEntry = newRequest(logEntry)
		ctx := context.WithValue(r.Context(), helpers.LoggerString, logEntry)

		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			err := htmlTemplates.GetTmpls().ExecuteTemplate(w, "checkLink.html", htmlTemplates.CheckLinkPageData{})
			if err != nil {
				logEntry.Error("Error sending error", err)
			}
		}
		a := r.Form

		queries := r.URL.Query()
		if _, ok := queries["link"]; !ok {
			queries["link"] = []string{a.Get("link")}
		}

		if _, ok := queries["link"]; !ok {
			w.WriteHeader(http.StatusOK)
			err := htmlTemplates.GetTmpls().ExecuteTemplate(w, "checkLink.html", htmlTemplates.CheckLinkPageData{})
			if err != nil {
				logEntry.Error("Error sending error", err)
			}
			return
		}
		if is, err := checkLink(queries["link"][0]); err != nil || !is {
			w.WriteHeader(http.StatusOK)
			err := htmlTemplates.GetTmpls().ExecuteTemplate(w, "checkLink.html", htmlTemplates.CheckLinkPageData{})
			if err != nil {
				logEntry.Error("Error sending error", err)
			}
			return
		} else {
			ctx = context.WithValue(ctx, helpers.LongLinkString, queries["link"][0])
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ManageShortLink(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logrus.New()
		logEntry := logrus.NewEntry(logger)
		logEntry = newRequest(logEntry)

		path := r.URL.Path[1:]
		shortPath, err := Base64ToInt(path)
		if err != nil {
			//http.Redirect(w, r, helpers.ShortLinkBase+"links", http.StatusSeeOther)
			w.WriteHeader(http.StatusNotFound)
			err = htmlTemplates.GetTmpls().ExecuteTemplate(w, "notfound.html", htmlTemplates.NotFoundPageData{})
			logEntry.Error("Error sending error: ", err)
		}

		ctx := context.WithValue(r.Context(), helpers.ShortLinkString, shortPath)
		ctx = context.WithValue(ctx, helpers.LoggerString, logEntry)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Base64ToInt(base string) (int64, error) {
	arr, err := b64.StdEncoding.DecodeString(base)
	if err != nil {
		return 0, err
	}
	val := int64(0)
	size := len(arr)
	for i := 0; i < size; i++ {
		*(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&val)) + uintptr(i))) = arr[i]
	}
	return val, nil
}
