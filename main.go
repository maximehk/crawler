package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/maximehk/crawler/download"
	"gopkg.in/xmlpath.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

func extractPics(content download.Response) (picUrls []string) {
	reader := bytes.NewReader(content.Data)
	xmlroot, err := xmlpath.ParseHTML(reader)
	if err != nil {
		log.Fatal(err)
	}

	path := xmlpath.MustCompile(`//img/@src`)
	iter := path.Iter(xmlroot)
	for iter.Next() {
		url := iter.Node().String()
		if strings.HasSuffix(url, ".jpg") {
			picUrls = append(picUrls, url)
		}
	}
	return
}

func createDir(dirPath string) {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.Mkdir(dirPath, 0700)
	}

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
	picUrls := extractPics(resp)

	// Enqueue URLs asynchronously and move on to processing the result
	go func() {
		for _, url := range picUrls {
			urls <- url
		}
	}()

	// Save the results to disk
  createDir("img")
	var wg sync.WaitGroup
	for i := 0; i < len(picUrls); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			resp := <-responses
			if resp.Error != nil {
				log.Println(resp.Error)
				return
			}
			filename := fmt.Sprintf("img/img_%d.jpg", i+1)
			err := ioutil.WriteFile(filename, []byte(resp.Data), 0644)
			if err != nil {
				log.Println(err)
			}
			fmt.Println("file saved to ", filename)
		}(i)

	}
	wg.Wait()
}
