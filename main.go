package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/maximehk/crawler/download"
	"gopkg.in/xmlpath.v2"
	"io/ioutil"
	"log"
	"strings"
	"sync"
)

func extractPics(body []byte) (picUrls []string) {
	reader := bytes.NewReader(body)
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
	var requestedUrl = flag.String("url", "", "url of a website with pics")
	flag.Parse()

	urls := make(chan string)
	responses := make(chan download.Response)
	go download.Downloader(urls, responses)

	// Download the requested page and extract the image URLs
	urls <- *requestedUrl
	resp := <-responses
	picUrls := extractPics(resp.Data)

	// Enqueue URLs asynchronously and move on to processing the result
	go func() {
		for _, url := range picUrls {
			urls <- url
		}
	}()

	// Save the results to disk
	var wg sync.WaitGroup
	for i := 0; i < len(picUrls); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			resp := <-responses
			filename := fmt.Sprintf("img/img_%d.jpg", i+1)
			err := ioutil.WriteFile(filename, []byte(resp.Data), 0644)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("file saved to ", filename)
		}(i)

	}
	wg.Wait()
}
