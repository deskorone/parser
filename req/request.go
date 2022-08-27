package req

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

type Hotel struct {
	Name   string  `json:"name"`
	Price  int     `json:"price"`
	Rating float32 `json:"rating"`
}

// Функция которая распределает запросы по горрутинам
func DoRequest(count int, arr []Hotel) error {
	lenght := len(arr)
	ch := make(chan Hotel)
	closeChan := make(chan int)

	defer close(ch)
	defer close(closeChan)

	var wg sync.WaitGroup
	gorutine := func() {
		defer wg.Done()
		for {
			select {
			case h := <-ch:
				e := makeReq(h)
				if e != nil {
					fmt.Println(e.Error())
				}
			case <-closeChan:
				return
			}
		}
	}
	for i := 0; i < count; i++ {
		wg.Add(1)
		go gorutine()
	}
	for i, h := range arr {
		ch <- h
		if i == lenght-1 {
			for j := 0; j < count; j++ {
				closeChan <- 0
			}
		}
	}
	wg.Wait()
	return nil
}

// Функция которая делает запрос на сервер
func makeReq(h Hotel) error {
	body, err := json.Marshal(h)
	if err != nil {
		return err
	}
	rbody := bytes.NewBuffer(body)

	resp, err := http.Post("http://localhost:8080/save", "application/json", rbody)
	if err != nil {
		return err
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	sb := string(body)
	log.Printf(sb)
	err = resp.Body.Close()

	return err
}
