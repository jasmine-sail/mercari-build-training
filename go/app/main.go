package main

import (
	"crypto/sha256"
	"eccoding/hex"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"encoding/json"

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

//Iten構造体　JSONオブジェクト内のキーが name になる
type Item struct{
	Name string `json:"name"`
	Category string `json:"category"`
	Image string `json:"image"`
}
func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	imageFile,err := c.FormFile("image")

	item := Item{name,category}
	byte,err := json.Marshal(item)
	if err != nil{
		return err
	}
	jsonBytes := []byte()(`{"name":name,"category":category}`)
	var item Item
	if err := json.Unmarshal(jsonBytes,&item); err := nil {
		return err
	}
		if err != nil {
			return err
		}
	c.Logger().Infof("Receive item: %s, Category: %s", name, category)
	


	file,err := os.ReadFile("items.json")
	if err != nil && !os.IsNotExist(err){
		return err
	}
	defer file.Close()

	 encoder := json.NewEncoder(file)
	 if err := encoder.Encode(item);err != nil {
		return err
	 }

	
	//画像の保存
	imagePath, err := saveImage(imageFile)
	if err != nil{
		return err
	}
	
	item := Item{
		Name: name,
		Category: category,
		Image: imagePath,
	}

	items = append(items,Item)

	newItemsJSON, err := json.Marshal(map[string][]item{"items": items})
	if err != nil {
		return err
	}

	if err := os.WriteFile("items.json", newItemsJSON,0644); err != nil {
		return err
	}



	message := fmt.Sprintf("item received: %s, Category: %s, Image: %s", name, category,imagePath)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
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

	frontURL := os.Getenv("FRONT_URL")
	if frontURL == "" {
		frontURL = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontURL},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem)  //ブラウザはGETでリクエストしているのにPOSTなので返ってこない
	e.GET("/image/:imageFilename", getImg)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
