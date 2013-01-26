package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"komrade/store"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	port   = flag.String("port", "8000", "Port for http server to bind.")
	jobPat = regexp.MustCompile(`\A\/jobs(\/(.*))?$`)
)

func init() {
	flag.Parse()
}

func main() {
	http.HandleFunc("/", router)
	err := http.ListenAndServe(":"+*port, nil)
	if err != nil {
		fmt.Printf("at=error error=\"Unable to start http server.\"\n")
		os.Exit(1)
	}
}

func router(w http.ResponseWriter, r *http.Request) {
	if jobPat.MatchString(r.URL.Path) {
		switch r.Method {
		case "PUT":
			handleNewJob(w, r)
			return
		case "GET":
			handleGetJob(w, r)
			return
		case "DELETE":
			handleDeleteJob(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

func handleNewJob(w http.ResponseWriter, r *http.Request) {
	token, err := ParseToken(r)
	if err != nil {
		writeJsonErr(w, 401, err)
		return
	}

	matches := jobPat.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 || len(matches[1]) == 0 {
		writeJsonErr(w, 400, errors.New("Job id not found."))
		return
	}
	// Trim the forward slash that came from the url path.
	id := string(matches[1][1:])
	if len(id) == 0 {
		writeJsonErr(w, 400, errors.New("Job id not found."))
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeJsonErr(w, 500, err)
		return
	}

	j := store.Job{Id: id, QueueId: token}
	err = json.Unmarshal(b, &j.Payload)
	if err != nil {
		writeJsonErr(w, 400, err)
		return
	}

	if ok := j.Get(); ok {
		WriteJson(w, 200, j)
		return
	}

	err = j.Insert()
	if err != nil {
		writeJsonErr(w, 500, err)
		return
	}
	WriteJson(w, 201, j)
}

func handleGetJob(w http.ResponseWriter, r *http.Request) {
	token, err := ParseToken(r)
	if err != nil {
		writeJsonErr(w, 401, err)
		return
	}

	q := r.URL.Query()
	limit := q.Get("limit")
	if len(limit) == 0 {
		limit = "1"
	}

	jobs, err := store.GetJobs(token, limit)
	if err != nil {
		writeJsonErr(w, 500, err)
		return
	}
	WriteJson(w, 200, jobs)
}

func handleDeleteJob(w http.ResponseWriter, r *http.Request) {
	matches := jobPat.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 || len(matches[1]) == 0 {
		writeJsonErr(w, 400, errors.New("Job id not found."))
		return
	}
	// Trim the forward slash that came from the url path.
	id := string(matches[1][1:])
	if len(id) == 0 {
		writeJsonErr(w, 400, errors.New("Job id not found."))
		return
	}
	job := store.Job{Id: id}
	if err := job.Delete(); err != nil {
		writeJsonErr(w, 400, err)
		return
	}
	WriteJson(w, 200, job)
}

func WriteJson(w http.ResponseWriter, status int, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		writeJsonErr(w, 500, err)
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		w.Write(b)
		w.Write([]byte("\n"))
	}
}

func writeJsonErr(w http.ResponseWriter, status int, err error) {
	fmt.Println("error:", err)
	WriteJson(w, status, map[string]string{"message": "Internal server error"})
}

func ParseToken(r *http.Request) (string, error) {
	header, ok := r.Header["Authorization"]
	if !ok {
		return "", errors.New("Authorization header not set.")
	}

	auth := strings.SplitN(header[0], " ", 2)
	if len(auth) != 2 {
		return "", errors.New("Malformed header.")
	}

	userPass, err := base64.StdEncoding.DecodeString(auth[1])
	if err != nil {
		return "", errors.New("Malformed encoding.")
	}

	parts := strings.Split(string(userPass), ":")
	if len(parts) != 2 {
		return "", errors.New("Password not supplied.")
	}

	return parts[1], nil
}
