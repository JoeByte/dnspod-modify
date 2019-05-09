package main

import (
	"fmt"
	"log"
	"time"
	"strings"
	"regexp"
	"net/url"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

const dnsId = ""
const dnsToken = ""
const dnsEndpoint = "https://dnsapi.cn/"

var currentIp = ""

type status struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type domain struct {
	Id     string
	Name   string
	Status string
}
type record struct {
	Id     string
	Name   string
	Value  string
	Status string
	Line   string
	Type   string
	Mx     string
	Mtime  string `json:"updated_on"`
}
type response struct {
	Status  status
	Domain  domain
	Records []record
}

func main() {
	var (
		domain = "xxtime.com"
		sub    = "www"
	)

	for {
		ip := getIpAddr()
		isIpAddr, _ := regexp.MatchString("^[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}$", ip)
		if isIpAddr && currentIp != ip {
			currentIp = ip
			d, r := getRecord(domain, sub)
			modifyRecord(sub, ip, d, r)
			fmt.Printf("set ip %s\n", ip)
		}
		fmt.Printf("current ip %s\n", ip)
		time.Sleep(600 * time.Second)
	}
}

func modifyRecord(sub string, ip string, d domain, r record) bool {
	uri := "Record.Modify"
	var argv = make(map[string]string)
	argv["domain_id"] = d.Id
	argv["record_id"] = r.Id
	argv["sub_domain"] = sub
	argv["record_type"] = "A"
	argv["record_line"] = "默认"
	argv["value"] = ip
	argv["ttl"] = "600"
	argv["status"] = "enable"
	requestApi(uri, "POST", argv)
	return true
}

func getRecord(domain string, sub string) (domain, record) {
	uri := "Record.List"
	var argv = make(map[string]string)
	argv["domain"] = domain
	if sub != "" {
		argv["sub_domain"] = sub
	}
	result, err := requestApi(uri, "POST", argv)
	if err != nil {
		log.Println(err)
	}
	r := result.(response)
	return r.Domain, r.Records[0]
}

func requestApi(uri string, method string, argv map[string]string) (interface{}, error) {
	var (
		c        = &http.Client{Timeout: time.Second * 10,}
		httpAddr = dnsEndpoint + uri
		data     = url.Values{}
		header   = make(map[string]string)
	)
	header["Content-Type"] = "application/x-www-form-urlencoded"
	argv["login_token"] = dnsId + "," + dnsToken
	argv["format"] = "json"
	for k, v := range argv {
		data.Set(k, v)
	}
	req, err := http.NewRequest(method, httpAddr, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		req.Header.Set(string(k), v)
	}

	resp, _ := c.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result response
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Println(err)
	}
	return result, nil
}

// @see https://www.ipify.org/
// @see https://seeip.org/
// @see cip.cc
func getIpAddr() string {
	url := "https://api.ipify.org"
	req, _ := http.NewRequest("GET", url, nil)
	client := &http.Client{Timeout: time.Second * 10,}
	res, _ := client.Do(req)
	body, _ := ioutil.ReadAll(res.Body)
	return string(body)
}
