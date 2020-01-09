package main

import (
	"fmt"
	"goddns/alidns"
	"goddns/dnspod"
	"log"
	"net/http"
)

func aliyunDDNS(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //解析参数，默认是不会解析的

	key := r.Form.Get("key")
	secret := r.Form.Get("secret")
	domain := r.Form.Get("domain")
	record := r.Form.Get("record")
	currentIP := r.Form.Get("ip")

	aliDNS := alidns.NewAliDNS(key, secret)
	records := aliDNS.GetDomainRecords(domain, record)

	if records == nil || len(records) == 0 {
		log.Printf("Cannot get subdomain %s from AliDNS.\r\n", record)
		fmt.Fprintf(w, "0")
		return
	}

	if records[0].Value != currentIP {
		records[0].Value = currentIP
		if err := aliDNS.UpdateDomainRecord(records[0]); err != nil {
			log.Printf("Failed to update IP for subdomain:%s\r\n", record)
			fmt.Fprintf(w, "0")
		} else {
			log.Printf("IP updated for subdomain:%s\r\n", record)
			fmt.Fprintf(w, "1")
		}
	} else {
		fmt.Fprintf(w, "1")
	}
}

func dnspodDDNS(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //解析参数，默认是不会解析的
	token := r.Form.Get("token")
	domain := r.Form.Get("domain")
	record := r.Form.Get("record")
	currentIP := r.Form.Get("ip")

	dnspodDNS := dnspod.NewDnspod(token)

	domainID := dnspodDNS.GetDomain(domain)

	if domainID != -1 {
		subDomainID, oldIP := dnspodDNS.GetSubDomain(domainID, record)
		if oldIP != currentIP {
			log.Printf("Begin Update %s ip\r\n", record)
			dnspodDNS.UpdateIP(domainID, subDomainID, record, currentIP)
			fmt.Fprintf(w, "1")
		} else {
			log.Println("Do not need update")
			fmt.Fprintf(w, "1")
		}
	} else {
		fmt.Fprintf(w, "0")
	}
}

func main() {
	http.HandleFunc("/aliyun", aliyunDDNS)   //设置访问的路由
	http.HandleFunc("/dnspod", dnspodDDNS)   //设置访问的路由
	err := http.ListenAndServe(":7000", nil) //设置监听的端口
	// err := http.ListenAndServeTLS(":7000", "certificate.crt", "certificate.key", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
