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
		*goroutineCount = 1
	}
	str, err := getDocFromWebSite(*url)

	if err != nil {
		return
	}

	arr, err := parseDocument(str)
	if err != nil {
		fmt.Printf("err: %v\n", err.Error())
		return
	}

	err = req.DoRequest(*goroutineCount, arr)
	if err != nil {
		return
	}

}

//Функция которая берет документ с сайта
func getDocFromWebSite(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(res.Body)
	return string(body), err
}

// Функция которая парсит документ
func parseDocument(str string) ([]req.Hotel, error) {

	doc, err := html.Parse(strings.NewReader(str))

	if err != nil {
		return nil, err
	}

	var foo func(n *html.Node, arr *[]req.Hotel)
	var h req.Hotel

	foo = func(n *html.Node, arr *[]req.Hotel) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == NAMEURL_CLASS {
					h.Name = n.FirstChild.Data
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
						h.Rating = 0
						break
					}
					h.Rating = float32(rating)
					break
				}
			}
		}
		if n.Type == html.ElementNode && n.Data == "span" {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == PRICE_CLASS {
					// Перепробовал все способы строка не тримится и не сплитится пришлось делать черещ костыль (с тем символом тоде все перепробывал)
					f := strings.TrimSuffix(n.FirstChild.Data, "₽")
					f = strings.TrimSpace(f)
					array := make([]int, 0)
					for _, i := range f {
						n, err := strconv.Atoi(string(i))
						if err != nil {
							continue
						}
						array = append(array, n)
					}

					num := toInt(array)

					h.Price = num
					*arr = append(*arr, h)
					break
				}
				if a.Key == "class" && a.Val == NAME_CLASS {
					h.Name = h.Name + " " + n.FirstChild.Data
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			foo(c, arr)
		}
	}

	arr := make([]req.Hotel, 0)

	foo(doc, &arr)

	// убираю первый элемент потому что он в нем нет информации об отеле
	arr = removeByIndex(arr, 0)
	return arr, err
}

//Функция удаляющая элемент по индексу из массива
func removeByIndex(arr []req.Hotel, i int) []req.Hotel {
	if len(arr) > 0 {
		arr[i] = arr[len(arr)-1]
		return arr[:len(arr)-1]
	}
	return arr
}

//Функция делающая сичло из массива чисел
func toInt(arr []int) int {
	res := 0
	op := 1
	for i := len(arr) - 1; i >= 0; i-- {
		res += arr[i] * op
		op *= 10
	}
	return res
}
