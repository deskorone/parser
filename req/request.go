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

// DoRequest Функция которая распределает запросы по горрутинам
func DoRequest(count int, arr []Hotel) error {
	ch := make(chan Hotel)
	// Канал для уведомления горутин о завершени работы
	closeChan := make(chan int)

	defer close(ch)
	defer close(closeChan)

	var wg sync.WaitGroup
	// Функция которая при создании слушает ранее созданные каналы ch closeChan
	// Можно сделать отдельной функцией, но тогда нужно передавать ссылку на waitgroup и канал
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
	for _, h := range arr {
		ch <- h
	}
	for j := 0; j < count; j++ {
		closeChan <- 0
	}
	wg.Wait()
	return nil
}

// Функция которая делает запрос на сервер
func makeReq(hotel Hotel) error {
	body, err := json.Marshal(hotel)
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
	stringBody := string(body)
	log.Printf(stringBody)
	err = resp.Body.Close()

	return err
}
