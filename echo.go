package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
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

const tpl = `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>UPYUN Notify Receiver</title>
<link rel="stylesheet" href="http://static-cache.b0.upaiyun.com/css/bootstrap.min.css">
<link rel="stylesheet" href="http://static-cache.b0.upaiyun.com/css/bootstrap-theme.min.css">
<script src="http://static-cache.b0.upaiyun.com/js/jquery.min.js"></script>
<script src="http://static-cache.b0.upaiyun.com/js/bootstrap.min.js"></script>
<style type="text/css">
.bs-example{
    margin: auto;
    width: 1000px;
}
</style>
</head><body><div class="bs-example">
<table class="table table-bordered table-striped">
        <thead>
            <tr>
                <th>Date</th>
                <th>Method</th>
				<th>RequestURI</th>
                <th>RemoteAddr</th>
                <th>UserAgent</th>
				<th>Body</th>
            </tr>
        </thead>
        <tbody>
		{{range .}}
        <tr class="success"> 
            <td>{{.Date}}</td>
            <td>{{.Method}}</td>
            <td>{{.RequestURI}}</td>
            <td>{{.RemoteAddr}}</td>
            <td>{{.UA}}</td>
            <td>{{.Body}}</td>
        </tr>
		{{end}}
</tbody></table></div></body></html>
`

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
		req.Body = string(b)
	}

	if err != nil {
		req.Body = fmt.Sprint(err)
	}

	update(req)
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	t := template.New("test")
	t, _ = t.Parse(tpl)
	t.Execute(w, Reqs)
}

func main() {
	num := flag.Int("r", 50, "max request number to show")
	port := flag.String("p", "1001", "server listen port")
	flag.Parse()
	maxReqNum = *num
	http.HandleFunc("/echo", echoHandler)
	http.HandleFunc("/query", queryHandler)
	http.ListenAndServe(":"+*port, nil)
}
