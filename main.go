package main

import (
	"fmt"
	"github.com/maximehk/crawler/download"
	"gopkg.in/xmlpath.v2"
	"io/ioutil"
	"log"
	"strings"
	"sync"
)

func extractPics(body string) (picUrls []string) {
	reader := strings.NewReader(body)
	xmlroot, xmlerr := xmlpath.ParseHTML(reader)
	if xmlerr != nil {
		log.Fatal(xmlerr)
	}

	var xpath string
	xpath = `//img/@src`
	path := xmlpath.MustCompile(xpath)
	iter := path.Iter(xmlroot)
	for iter.Next() {
		url := iter.Node().String()
		if strings.HasSuffix(url, ".jpg") {
			picUrls = append(picUrls, url)
		}
	}
	return
}

func main() {
	testUrls := [1]string{"http://...url of a website with pics.net/"}
	urls := make(chan string, 99)
	responses := make(chan download.Response)
	go download.Downloader(urls, responses)

	for _, url := range testUrls {
		urls <- url
	}

	var picUrls []string

	for i := 0; i < len(testUrls); i++ {
		resp := <-responses
		picUrls = append(picUrls, extractPics(resp.Data)...)
	}

	for _, url := range picUrls {
		urls <- url
	}

	var wg sync.WaitGroup
	for i := 0; i < len(picUrls); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			resp := <-responses
			filename := fmt.Sprintf("img/img_%d.jpg", i+1)
			ioutil.WriteFile(filename, []byte(resp.Data), 0644)
			fmt.Println("file saved to ", filename)
		}(i)

	}
	wg.Wait()

}
