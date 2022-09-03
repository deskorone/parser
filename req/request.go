package req

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type Hotel struct {
	Name   string  `json:"name"`
	Price  int     `json:"price"`
	Rating float32 `json:"rating"`
}

// DoRequest Функция которая распределает запросы по горрутинам
func DoRequest(count int, arr []Hotel) error {
	ch := make(chan Hotel)

	// Контекст для завершения горутин
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	// Функция которая при создании слушает ранее созданные каналы ch closeChan
	gorutine := func() {
		defer wg.Done()
		for {
			select {
			case h := <-ch:
				e := makeReq(h)
				if e != nil {
					fmt.Println(e.Error())
				}
			case <-ctx.Done():
				return
			}
		}
	}
	for i := 0; i < count; i++ {
		wg.Add(1)
		go gorutine()
	}
	for _, h := range arr {
		ch <- h
	}
	cancel()
	wg.Wait()
	return nil
}

// Функция которая делает запрос на сервер
func makeReq(hotel Hotel) error {
	body, err := json.Marshal(hotel)
	if err != nil {
		return err
	}

	time.Sleep(700 * time.Millisecond)
	rbody := bytes.NewBuffer(body)

	resp, err := http.Post("http://localhost:8080/save", "application/json", rbody)
	if err != nil {
		return err
	}
	body, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalln(err)
	}
	stringBody := string(body)
	log.Printf(stringBody)
	err = resp.Body.Close()

	return err
}
