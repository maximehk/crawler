package download

import (
	"bytes"
	"log"
	"net/http"
)

const MaxWorkers = 5

type Response struct {
	Url  string
	Data []byte
  Error error
}

func download(url string) (response Response) {
	log.Println("Downloading ", url)
	response.Url = url
	resp, err := http.Get(url)
	if err != nil {
    response.Error = err
    return
	}

	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	response.Data = buf.Bytes()
	log.Println("Downloaded ", url)
	return
}

func downloadWorker(urls <-chan string, responses chan<- Response) {
	for url := range urls {
		responses <- download(url)
	}
}

func Downloader(urls <-chan string, responses chan<- Response) {
	for i := 0; i < MaxWorkers; i++ {
		go downloadWorker(urls, responses)
	}
}
