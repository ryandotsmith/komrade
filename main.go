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
	port      = flag.String("port", "8000", "Port for http server to bind.")
	heartPat  = regexp.MustCompile(`\A\/jobs\/(.*)\/heartbeats$`)
	failPat   = regexp.MustCompile(`\A\/jobs\/(.*)\/failures\/(.*)$`)
	jobPat    = regexp.MustCompile(`\A\/jobs(\/(.*))?$`)
	delAllPat = regexp.MustCompile(`\A\/delete-all-jobs$`)
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
	if delAllPat.MatchString(r.URL.Path) {
		if r.Method == "POST" {
			handleDeleteAll(w, r)
			return
		}
	}
	if heartPat.MatchString(r.URL.Path) {
		if r.Method == "POST" {
			handleHeartBeat(w, r)
			return
		}
	}
	if failPat.MatchString(r.URL.Path) {
		if r.Method == "PUT" {
			handleFailedJob(w, r)
			return
		}
	}
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

func handleHeartBeat(w http.ResponseWriter, r *http.Request) {
	token, err := ParseToken(r)
	if err != nil {
		writeJsonErr(w, 401, err)
		return
	}

	matches := heartPat.FindStringSubmatch(r.URL.Path)
	if len(matches) != 2 {
		writeJsonErr(w, 400, errors.New("Missing job id or failured id."))
		return
	}
	id := string(matches[1])
	if len(id) == 0 {
		writeJsonErr(w, 400, errors.New("Job id not found."))
		return
	}

	job := store.Job{Id: id, QueueId: token}
	err = job.HeartBeat()
	if err != nil {
		writeJsonErr(w, 500, errors.New("Unable to commit heartbeat."))
		return
	}
	WriteJson(w, 201, map[string]string{"message": "OK"})
}

func handleDeleteAll(w http.ResponseWriter, r *http.Request) {
	token, err := ParseToken(r)
	if err != nil {
		writeJsonErr(w, 401, err)
		return
	}

	jobCount, err := store.DeleteAllJobs(token)
	if err != nil {
		writeJsonErr(w, 500, err)
	}

	WriteJson(w, 200, map[string]int64{"deleted": jobCount})
}

func handleFailedJob(w http.ResponseWriter, r *http.Request) {
	token, err := ParseToken(r)
	if err != nil {
		writeJsonErr(w, 401, err)
		return
	}

	matches := failPat.FindStringSubmatch(r.URL.Path)
	if len(matches) < 3 {
		writeJsonErr(w, 400, errors.New("Missing job id or failured id."))
		return
	}
	jid := string(matches[1])
	fid := string(matches[2])
	if len(jid) == 0 {
		writeJsonErr(w, 400, errors.New("Job id not found."))
		return
	}
	if len(fid) == 0 {
		writeJsonErr(w, 400, errors.New("Failure id not found."))
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeJsonErr(w, 500, err)
		return
	}

	f := store.FailedJob{Id: fid, JobId: jid, QueueId: token}

	if ok := f.Get(); ok {
		WriteJson(w, 200, f)
		return
	}

	err = json.Unmarshal(b, &f.Payload)
	if err != nil {
		writeJsonErr(w, 400, err)
		return
	}

	err = f.Insert()
	if err != nil {
		writeJsonErr(w, 500, err)
		return
	}
	WriteJson(w, 201, f)
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
	job := store.Job{Id: id, QueueId: token}
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
