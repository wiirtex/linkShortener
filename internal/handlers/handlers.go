package handlers

import (
	"encoding/base64"
	"linkShortener/internal/memory"
	"linkShortener/internal/memory/cache"
	"linkShortener/pkg/helpers"
	htmlTemplates "linkShortener/templates"
	"net/http"
	"time"
	"unsafe"

	"github.com/sirupsen/logrus"
)

func AddLink(w http.ResponseWriter, r *http.Request) {
	cache := cache.GetDBCacheInstance()

	long, ok := r.Context().Value(helpers.LongLinkString).(string)
	if !ok || long == "" {
		r.Context().Value(helpers.LoggerString).(*logrus.Entry).Error("Error not string (or empty) longLink is in handler", r.Context())
		w.WriteHeader(http.StatusInternalServerError)
		err := htmlTemplates.GetTmpls().ExecuteTemplate(w, "error.html", htmlTemplates.ErrorPageData{
			Error: "not string (or empty) longLink is in handler",
		})
		if err != nil {
			r.Context().Value(helpers.LoggerString).(*logrus.Entry).Error("Error sending error", err)
		}
		return
	}

	memoryRequest := memory.MemoryRequest{
		Long:      long,
		CreatedAt: time.Now().UTC(),
		Author:    "test",
	}

	shortLink, err := cache.AddEntry(memoryRequest)
	if err != nil {
		r.Context().Value(helpers.LoggerString).(*logrus.Entry).Error("Error add to cache error ", err, memoryRequest)
		w.WriteHeader(http.StatusInternalServerError)
		err := htmlTemplates.GetTmpls().ExecuteTemplate(w, "error.html", htmlTemplates.ErrorPageData{
			Error: "server have some problems with cache",
		})
		if err != nil {
			r.Context().Value(helpers.LoggerString).(*logrus.Entry).Error("Error sending error", err)
		}
		return
	}
	memoryRequest.Short = shortLink
	w.WriteHeader(http.StatusOK)
	err = htmlTemplates.GetTmpls().ExecuteTemplate(w, "success.html", htmlTemplates.SuccessPageData{
		GeneratedLink: helpers.GetConfig().ShortLinkBase + intToBase64(memoryRequest.Short),
	})
	if err != nil {
		r.Context().Value(helpers.LoggerString).(*logrus.Entry).Error("Error sending correct link", err)
	}

}

func intToBase64(num int64) string {
	size := 0
	for i := 8; i <= 64; i += 8 {
		size += 1
		if num>>i == 0 {
			break
		}
	}
	arr := make([]byte, size)
	for i := 0; i < size; i++ {
		byt := *(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&num)) + uintptr(i)))
		arr[i] = byt
	}
	return base64.StdEncoding.EncodeToString(arr)
}

func RedirectLink(w http.ResponseWriter, r *http.Request) {
	path, ok := r.Context().Value(helpers.ShortLinkString).(int64)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		err := htmlTemplates.GetTmpls().ExecuteTemplate(w, "error.html", htmlTemplates.ErrorPageData{
			Error: "some mistake in work of the server",
		})
		r.Context().Value(helpers.LoggerString).(*logrus.Entry).Error("Some error with shortLink", err, path)
		return
	}
	if path == 0 {
		http.Redirect(w, r, helpers.GetConfig().ShortLinkBase+"links", http.StatusSeeOther)
		return
	} else {

		memoryRequest := memory.MemoryRequest{
			Short: path,
		}

		cache := cache.GetDBCacheInstance()

		if memoryResponse, err := cache.GetEntry(memoryRequest); err == nil { // cache hit
			http.Redirect(w, r, memoryResponse.Long, 301)
			return
		} else {
			w.WriteHeader(http.StatusNotFound)
			err := htmlTemplates.GetTmpls().ExecuteTemplate(w, "notfound.html", htmlTemplates.NotFoundPageData{})
			if err != nil {
				r.Context().Value(helpers.LoggerString).(*logrus.Entry).Error("Error sending 404", err, path)
			}
		}
	}
}
