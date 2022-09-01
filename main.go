package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"urlParser/req"

	// "sync"

	"golang.org/x/net/html"
)

// Конcтанты классов в которых находится информация о отелях
const (
	NAMEURL_CLASS = "Link Link_theme_normal Link_view_default STEEy Y_QIb Link_lego KlhYG" //класс отвечаюзий за название отеля
	NAME_CLASS    = "d2q3W"
	PRICE_CLASS   = "Akpkj"
	RATING_CLASS  = "FlrxR u8BKo kWZoP LyIFS"
)

func main() {
	url := flag.String("path", "https://travel.yandex.ru/hotels/moscow/", "hotels url")
	goroutineCount := flag.Int("r", 1, "hotels url")
	flag.Parse()
	if *goroutineCount < 1 {
		fmt.Println("Вы ввели некоректное колличество горутин (")
		return
	}
	docFromWebSite, err := getDocFromWebSite(*url)

	if err != nil {
		fmt.Println("Прооизощла ошибка при получении документа, проверьте свое соединение!")
		return
	}

	hotels, err := parseDocument(docFromWebSite)
	if err != nil {
		fmt.Printf("err: %v\n", err.Error())
		return
	}

	err = req.DoRequest(*goroutineCount, hotels)
	if err != nil {
		return
	}

}

//Функция которая берет документ с сайта
func getDocFromWebSite(url string) (string, error) {
	result, err := http.Get(url)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(result.Body)
	return string(body), err
}

// Функция которая парсит документ
func parseDocument(str string) ([]req.Hotel, error) {

	doc, err := html.Parse(strings.NewReader(str))

	if err != nil {
		return nil, err
	}

	var hotel req.Hotel

	// объявляю здесь иначе функция не будет работать рекурсивно
	var parseFunction func(n *html.Node, hotels *[]req.Hotel)

	// Функция которая парсит документ
	parseFunction = func(n *html.Node, hotels *[]req.Hotel) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == NAMEURL_CLASS {
					hotel.Name = n.FirstChild.Data
					break
				}
			}
		}
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == RATING_CLASS {
					rating, err := strconv.ParseFloat(
						strings.ReplaceAll(n.FirstChild.Data, ",", "."), 32)
					if err != nil {
						hotel.Rating = 0
						break
					}
					hotel.Rating = float32(rating)
					break
				}
			}
		}
		if n.Type == html.ElementNode && n.Data == "span" {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == PRICE_CLASS {
					// Перепробовал все способы строка не тримится и не сплитится пришлось делать черещ костыль (с тем символом тоде все перепробывал)
					word := strings.TrimSuffix(n.FirstChild.Data, "₽")
					word = strings.TrimSpace(word)
					numArray := make([]int, 0)
					for _, i := range word {
						n, err := strconv.Atoi(string(i))
						if err != nil {
							continue
						}
						numArray = append(numArray, n)
					}
					num := toInt(numArray)
					hotel.Price = num
					// Добавляю отель здесь потому что этот класс парсится последним
					*hotels = append(*hotels, hotel)
					break
				}
				if a.Key == "class" && a.Val == NAME_CLASS {
					hotel.Name = hotel.Name + " " + n.FirstChild.Data
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parseFunction(c, hotels)
		}
	}

	hotels := make([]req.Hotel, 0)

	parseFunction(doc, &hotels)

	// убираю первый элемент потому что он в нем нет информации об отеле
	hotels = removeByIndex(hotels, 0)
	return hotels, err
}

//Функция удаляющая элемент по индексу из массива
func removeByIndex(arr []req.Hotel, i int) []req.Hotel {
	if len(arr) > 0 {
		arr[i] = arr[len(arr)-1]
		return arr[:len(arr)-1]
	}
	return arr
}

//Функция делающая число из массива чисел
func toInt(arr []int) int {
	res := 0
	op := 1
	for i := len(arr) - 1; i >= 0; i-- {
		res += arr[i] * op
		op *= 10
	}
	return res
}
