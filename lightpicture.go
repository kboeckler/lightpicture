package main

import (
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/studio-b12/gowebdav"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	closer := setupLogger()
	defer closer.Close()

	log.Infoln("Lightpicture is starting")
	server := setupServer()
	server.startServer()
}

func setupLogger() io.Closer {
	logWriter, err := os.OpenFile("log.json",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(logWriter)

	return logWriter
}

type server struct {
	cfg *config
}

func setupServer() server {
	cfg := readConfig()
	return server{cfg}
}

func (s server) getPicture(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(401)
		w.Write([]byte("required basic auth"))
		return
	}
	requestedWidth := 1280
	requestedHeight := 720
	var err error
	rw := r.URL.Query().Get("width")
	if len(rw) > 0 {
		requestedWidth, err = strconv.Atoi(rw)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("wrong width"))
			return
		}
	}
	rh := r.URL.Query().Get("height")
	if len(rh) > 0 {
		requestedHeight, err = strconv.Atoi(rh)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("wrong height"))
			return
		}
	}
	client := gowebdav.NewClient(fmt.Sprintf("%s%s/files/%s/", s.cfg.BaseUrl, s.cfg.HomePath, username), username, password)
	client.SetTimeout(5 * time.Second)
	requestedFile := r.URL.Path
	resultStream, err := client.ReadStream(requestedFile)
	if err != nil {
		pathError := err.(*os.PathError)
		statusError, isStatusError := pathError.Err.(gowebdav.StatusError)
		if isStatusError {
			w.WriteHeader(statusError.Status)
			return
		}
		log.Errorf("Error reading image %s: %v", requestedFile, err)
		w.WriteHeader(500)
		return
	}
	defer func(resultStream io.ReadCloser) {
		err := resultStream.Close()
		if err != nil {
			log.Warnf("Error closing file read stream from webdav: %v", err)
		}
	}(resultStream)
	image, err := imaging.Decode(resultStream)
	if err != nil {
		log.Errorf("Error decoding image %s: %v", requestedFile, err)
		w.WriteHeader(500)
		return
	}
	if float64(image.Bounds().Dx())/float64(image.Bounds().Dy()) < float64(requestedWidth)/float64(requestedHeight) {
		image = imaging.Resize(image, 0, requestedHeight, imaging.Lanczos)
	} else {
		image = imaging.Resize(image, requestedWidth, 0, imaging.Lanczos)
	}
	err = imaging.Encode(w, image, imaging.JPEG)
	if err != nil {
		log.Errorf("Error encoding image %s: %v", requestedFile, err)
		w.WriteHeader(500)
		return
	}
}

func (s server) openApi(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "openapi.html")
}

func (s server) startServer() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Path("/openapi/").HandlerFunc(s.openApi)
	myRouter.PathPrefix("/").HandlerFunc(s.getPicture)
	address := fmt.Sprintf("%s:%d", s.cfg.Hostname, s.cfg.Port)
	srv := &http.Server{
		Handler: myRouter,
		Addr:    address,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	if s.cfg.SSL {
		log.Infof("Listening on %s\n", "https://"+address)
		log.Fatal(srv.ListenAndServeTLS(s.cfg.CertFile, s.cfg.KeyFile))
	} else {
		log.Infof("Listening on %s\n", "http://"+address)
		log.Fatal(srv.ListenAndServe())
	}
}
