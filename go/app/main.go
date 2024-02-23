package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

// 構造体
type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

type ItemList struct {
	Items []Item `json:"items"`
}

// e.GET("/", root)
func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

// e.POST("/items", addItem) これでjsonファイルに追加！
func addItem(c echo.Context) error {
	// Get form data
	var itemlist ItemList
	var item Item
	item.Name = c.FormValue("name")
	item.Category = c.FormValue("category")
	c.Logger().Infof("Receive item: %s, %s", item.Name, item.Category)

	//4.jsonファイルの読み込み
	file, err := os.OpenFile("items.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}
	defer file.Close()

	//5.jsonファイルをdecode jsonのNewDecoderとDecode関数を使う
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&itemlist); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}

	// 6. step5でdecodeしたitemをstep3のitemに追加する
	itemlist.Items = append(itemlist.Items, item)

	// 7. jsonの書き込み用にファイルを開く
	//fileに再代入なので:=ではなく=で書く
	file, err = os.Create("items.json")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}
	defer file.Close()
	//8. jsonファイルに書き込み
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(itemlist); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}

	message := fmt.Sprintf("item received: %s,%s", item.Name, item.Category)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

// e.GET("/items",getItem)jsonファイルからデータを持ってくる！
func getItem(c echo.Context) error {
	file, err := os.Open("items.json")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}
	defer file.Close()
	var getitem ItemList
	if err := json.NewDecoder(file).Decode(&getitem); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}
	defer file.Close()

	return c.JSON(http.StatusOK, getitem)
}

func getImg(c echo.Context) error {
	// Create image path
	//画像ファイルが保存されているディレクトリのパスImgDirとクライアントから送信された画像ファイルの名前
	//これを結合して画像ファイルのパスを生成
	// ImgDir = /images,e.GET("/image/:imageFilename", getImg)より
	//imgPathは　/images/:imageFilenameになる
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))
	//拡張子がjpgかの確認
	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	//c.File()関数は、指定されたファイルパスに対応するファイルを
	//クライアントに送信するためのレスポンスを作成
	return c.File(imgPath)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	front_url := os.Getenv("FRONT_URL")
	if front_url == "" {
		front_url = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{front_url},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem)
	e.GET("/items", getItem)
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
