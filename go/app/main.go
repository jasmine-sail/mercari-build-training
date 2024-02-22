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
	file, err := os.OpenFile("items.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}
	defer file.Close()
	// Get form data
	var itemlist ItemList
	var item Item
	item.Name = c.FormValue("name")
	item.Category = c.FormValue("category")
	c.Logger().Infof("Receive item: %s, %s", item.Name, item.Category)

	itemlist.Items = append(itemlist.Items, item)
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
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
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
