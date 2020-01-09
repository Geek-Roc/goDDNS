package dnspod

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

// Dnspod token
type Dnspod struct {
	LoginToken string
}

var (
	baseURL  = "https://dnsapi.cn"
	instance *Dnspod
	once     sync.Once
)

type status struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	CreateTime string `json:"created_at"`
}

type domainRecordsResp struct {
	Status  status    `json:"status"`
	Domains []domains `json:"domains"`
}

type domains struct {
	ID int `json:"id"`
}

type domainRecords struct {
	Status  status    `json:"status"`
	Records []records `json:"records"`
}

type records struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

type updateRecords struct {
	Status status `json:"status"`
	Record record `json:"record"`
}

type record struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// NewDnspod function creates instance of Dnspod and return
func NewDnspod(loginToken string) *Dnspod {
	once.Do(func() {
		instance = &Dnspod{
			LoginToken: loginToken,
		}
	})
	return instance
}

// GetDomain returns specific domain by name
func (d *Dnspod) GetDomain(name string) int {
	log.Println("GetDomain:", name)

	resp := &domainRecordsResp{}

	values := url.Values{}
	values.Add("type", "all")
	values.Add("offset", "0")
	values.Add("length", "20")

	body, err := d.postData("/Domain.List", values)

	if err != nil {
		log.Println("Failed to get domain list...")
		return -1
	}

	if err := json.Unmarshal(body, resp); err != nil {
		log.Println("GetDomainRecords error.")
		log.Println(err)
		return -1
	}

	if resp == nil || len(resp.Domains) == 0 {
		log.Println("GetDomainRecords error.", resp.Status.Message)
		return -1
	}

	log.Println("GetDomainRecords id: ", resp.Domains[0].ID)
	return resp.Domains[0].ID
}

// GetSubDomain returns subdomain by domain id
func (d *Dnspod) GetSubDomain(domainID int, name string) (string, string) {
	log.Println("GetSubDomain:", domainID, name)

	resp := &domainRecords{}

	values := url.Values{}
	values.Add("domain_id", strconv.Itoa(domainID))
	values.Add("offset", "0")
	values.Add("length", "1")
	values.Add("sub_domain", name)

	body, err := d.postData("/Record.List", values)

	if err != nil {
		log.Println("Failed to get domain list")
		return "", ""
	}

	if err := json.Unmarshal(body, resp); err != nil {
		log.Println("GetSubDomain error")
		log.Println(err)
		return "", ""
	}

	if resp == nil || len(resp.Records) == 0 {
		log.Println("GetSubDomain error.", resp.Status.Message)
		return "", ""
	}

	log.Println("GetSubDomain ip: ", resp.Records[0].Value)
	return resp.Records[0].ID, resp.Records[0].Value
}

// UpdateIP update subdomain with current IP
func (d *Dnspod) UpdateIP(domainID int, subDomainID string, subDomainName string, ip string) {
	value := url.Values{}
	value.Add("domain_id", strconv.Itoa(domainID))
	value.Add("record_id", subDomainID)
	value.Add("sub_domain", subDomainName)
	value.Add("record_type", "A")
	value.Add("record_line", "默认")
	value.Add("value", ip)

	body, err := d.postData("/Record.Modify", value)

	if err != nil {
		log.Println("Failed to update record to new IP!")
		log.Println(err)
		return
	}

	resp := &updateRecords{}
	if err := json.Unmarshal(body, resp); err != nil {
		log.Println("UpdateIP error")
		log.Println(err)
	}

	if resp.Status.Code == "1" {
		log.Println("New IP updated!")
	} else {
		log.Println("Change IP Failed.")
	}
}

func (d *Dnspod) postData(url string, content url.Values) ([]byte, error) {
	client := &http.Client{}

	if client == nil {
		return nil, errors.New("failed to create HTTP client")
	}

	values := d.generateHeader(content)
	req, _ := http.NewRequest("POST", baseURL+url, strings.NewReader(values.Encode()))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", fmt.Sprintf("GoDNS/0.1 (%s)", ""))

	response, err := client.Do(req)

	if err != nil {
		log.Println("Post failed...")
		log.Println(err)
		return nil, err
	}

	defer response.Body.Close()
	resp, _ := ioutil.ReadAll(response.Body)

	return resp, nil
}

// GenerateHeader generates the request header for DNSPod API
func (d *Dnspod) generateHeader(content url.Values) url.Values {
	header := url.Values{}
	if d.LoginToken != "" {
		header.Add("login_token", d.LoginToken)
	}
	header.Add("format", "json")
	header.Add("lang", "en")
	header.Add("error_on_empty", "no")

	if content != nil {
		for k := range content {
			header.Add(k, content.Get(k))
		}
	}

	return header
}
