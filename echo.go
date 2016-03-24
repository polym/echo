package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

type RequestInfo struct {
	Date       string
	Method     string
	RemoteAddr string
	RequestURI string
	UA         string
	Body       string
}

var (
	Reqs      []RequestInfo
	maxReqNum int
	lock      = new(sync.Mutex)
)

func update(req RequestInfo) {
	lock.Lock()
	defer lock.Unlock()
	if len(Reqs) == maxReqNum {
		Reqs = Reqs[0 : maxReqNum-1]
	}

	arr := []RequestInfo{req}
	Reqs = append(arr, Reqs...)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	req := RequestInfo{
		Method:     r.Method,
		RemoteAddr: r.RemoteAddr,
		RequestURI: r.RequestURI,
		UA:         r.UserAgent(),
	}

	req.Date = fmt.Sprint(time.Now().Format(time.UnixDate))

	defer r.Body.Close()

	URL, err := url.ParseRequestURI(r.RequestURI)
	if err == nil {
		q := URL.Query()
		if v := q.Get("sleep"); v != "" {
			if sec, err := strconv.Atoi(v); err == nil {
				time.Sleep(time.Duration(sec) * time.Second)
			}
		}
		if v := q.Get("close"); v == "true" {
			panic(r)
		}
		b, _ := ioutil.ReadAll(r.Body)
		var out bytes.Buffer
		if json.Indent(&out, b, "", "    ") == nil {
			req.Body = string(out.Bytes())
		} else {
			req.Body = string(b)
		}
	}

	if err != nil {
		req.Body = fmt.Sprint(err)
	}

	update(req)
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	t := template.New("tpl.html")
	t, _ = t.ParseFiles("tpl.html")
	t.Execute(w, Reqs)
}

func decodeHandler(w http.ResponseWriter, r *http.Request) {
	URL, err := url.ParseRequestURI(r.RequestURI)
	if err == nil {
		q := URL.Query()
		if v := q.Get("base64"); v != "" {
			b, _ := base64.StdEncoding.DecodeString(v)
			w.Write(b)
		}
	}
}

func docHandler(w http.ResponseWriter, r *http.Request) {
	repo := "https://github.com/upyun/docs.git"
	user := "ohara"
	bucket := "upyundocs"
	pass := ""
	tmpPath := "/tmp/repo"

	os.RemoveAll(tmpPath)
	defer os.RemoveAll(tmpPath)

	URL, err := url.ParseRequestURI(r.RequestURI)
	if err == nil {
		q := URL.Query()
		if v := q.Get("pass"); v != "" {
			pass = v
			gitCmd := fmt.Sprintf("git clone %s %s", repo, tmpPath)
			_, err = exec.Command("/bin/bash", "-c", gitCmd).Output()
			if err == nil {
				upxCmd := fmt.Sprintf("cd %s && mkdocs build && cd site && upx login %s %s %s && upx put ./",
					tmpPath, bucket, user, pass)
				fmt.Println(upxCmd)
				b, err := exec.Command("/bin/bash", "-c", upxCmd).Output()
				fmt.Println(string(b), err)
			}
		}
	}

	fmt.Println(err)
}

func main() {
	num := flag.Int("r", 50, "max request number to show")
	port := flag.String("p", "1001", "server listen port")
	flag.Parse()
	maxReqNum = *num
	http.HandleFunc("/echo", echoHandler)
	http.HandleFunc("/query", queryHandler)
	http.HandleFunc("/doc", docHandler)
	http.HandleFunc("/decode", decodeHandler)
	http.ListenAndServe(":"+*port, nil)
}
