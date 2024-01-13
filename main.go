package main

import (
	"context"
	"net/http"
	URL "net/url"
	"log"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"github.com/mmpx12/optionparser"
)

var (
	client *http.Client
	proxies []string
	wg sync.WaitGroup
	proxyCheck []string
)

func proxyTest(scheme string, proxy string, urlTarget string, timeout string) {
	timeouts, _ := strconv.Atoi(timeout)
	proxyUrl, _ := URL.Parse(scheme + proxy)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*time.Duration(timeouts))
	client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}
	req, err := http.NewRequestWithContext(ctx, "GET", urlTarget, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return
	}
	fmt.Println(proxy)
	proxies = append(proxies, proxy)
	return
}

func writeFile(fileName string) {
	f, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	fmt.Println(proxies)
	for _, proxy := range proxies {
		_, err := f.WriteString(proxy + "\n")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func readFile(fileName string) string{
	data, err := os.ReadFile(fileName)
	if err != nil {
		return ""
	}
	
	return string(data)

}
func main() {
	var urlTarget, fileName, output, timeout string
	op := optionparser.NewOptionParser()
	op.On("-u", "--url TARGET", "url testing proxy ", &urlTarget)
	op.On("-o", "--output FILENAME", "filename output", &output)
	op.On("-t", "--timeout TIMEOUT", "timeout request", &timeout)
	op.On("-f", "--file FILENAME", "file contain url url get proxy", &fileName)
	err := op.Parse()
	if err != nil {
		return
	}
	fmt.Println(timeout)
	scheme := "http://"
	urls := strings.Split(readFile(fileName), "\n")
	for _, url := range urls {
		res, err := http.Get(url)
		if err != nil {
			continue
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		proxyList := strings.Split(string(body), "\n")
		fmt.Println(len(proxyList))
		proxyCheck = append(proxyCheck, proxyList...)
	}


	fmt.Println(len(proxyCheck))

	for _, ip := range proxyCheck {
		ip = strings.Replace(ip, "\r", "", -1)
		ip = ip
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			proxyTest(scheme, ip, urlTarget, timeout)
		}(ip)
	}
	wg.Wait()
	writeFile(output)
}