package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ImgDir  = "images"
	DB_PATH = "../db/mercari.sqlite3"
)

type Response struct {
	Message string `json:"message"`
}

// 構造体
type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Image    string `json:"image_name"`
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

	//var itemlist ItemList
	var item Item
	item.Name = c.FormValue("name")         //jacket
	item.Category = c.FormValue("category") //fashion
	imageFile, err := c.FormFile("image")   //imageファイル


	//画像ファイルの読み込み
	src, err := imageFile.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}
	defer src.Close()

	//hash化
	hash := sha256.New()
	if _, err := io.Copy(hash, src); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}
	//hash値の取得
	hashString := hex.EncodeToString(hash.Sum(nil))
	//hash化されたファイル名
	imageFilename := hashString + ".jpg"
	//ポストされた画像のファイルを <hash>.jpg(=imageFilename)という名前で保存
	savePath := filepath.Join(ImgDir, imageFilename)
	// /images/<hash>.jpgに新しいファイルを作る（ハッシュ化された画像の保存先）
	dst, err := os.Create(savePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}
	defer dst.Close()
	//src(読み込んだ画像ファイル)をdst(画像の保存先)に保存
	if _, err := io.Copy(dst, src); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}


	//item = Item{Name: item.Name, Category: item.Category, Image: imageFilename}


	// 6. step5でdecodeしたitemをstep3のitemに追加する
	//itemlist.Items = append(itemlist.Items, item)

	//データベースへの接続
	db, err := sql.Open("sqlite3", DB_PATH)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}
	defer db.Close()

	//商品の追加
	var categoryid int
	row := db.QueryRow("SELECT id FROM categories WHERE name = $1", item.Category)
	err = row.Scan(&categoryid)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err = db.Exec("INSERT INTO categories (name) VALUES ($1)", item.Category)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
			}
			row := db.QueryRow("SELECT id FROM categories WHERE name = $1", item.Category)
			err = row.Scan(&categoryid)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
			}
		} else {
			return c.JSON(http.StatusInternalServerError, Response{Message: "failed to insert category" + err.Error()}) //ここ
		}
	}
	_, err = db.Exec("INSERT INTO items (name, category_id, image_name) VALUES ($1, $2, $3)", item.Name, categoryid, imageFilename)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}

	//log
	//c.Logger().Infof("Receive item: %s, %s,%s", item.Name, item.Category, item.Image)
	//message := fmt.Sprintf("item received: %s,%s,%s", item.Name, item.Category, item.Image)
	c.Logger().Infof("Receive item: %s, %s,%s", item.Name, item.Category, imageFilename)
	message := fmt.Sprintf("item received: %s,%s,%s", item.Name, item.Category, imageFilename)


	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

// e.GET("/items",getItem)jsonファイルからデータを持ってくる！
func getItem(c echo.Context) error {
	//var item Item
	//データベースへの接続
	db, err := sql.Open("sqlite3", DB_PATH)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}
	defer db.Close()
	//データベースから商品の取得
	rows, err := db.Query("SELECT items.name, categories.name, items.image_name FROM items JOIN categories ON items.category_id = categories.id")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}
	defer rows.Close()

	var getitem ItemList
	for rows.Next() {
		var name, category, image string
		if err := rows.Scan(&name, &category, &image); err != nil {
			return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
		}
		items := Item{Name: name, Category: category, Image: image}
		getitem.Items = append(getitem.Items, items)
	}
	return c.JSON(http.StatusOK, getitem)
}

// e.GET("/image/:imageFilename", getImg)
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

// e.GET("/items/:id",getId )
func getId(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))


	//データベースへの接続
	db, err := sql.Open("sqlite3", DB_PATH)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}

	defer db.Close()


	rows, err := db.Query("SELECT items.name, categories.name, items.image_name FROM items JOIN categories ON items.category_id = categories.id WHERE items.id LIKE ?", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}
	defer rows.Close()
	var iditem ItemList
	for rows.Next() {
		var name, category, image string
		if err := rows.Scan(&name, &category, &image); err != nil {
			return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
		}
		items := Item{Name: name, Category: category, Image: image}
		iditem.Items = append(iditem.Items, items)
	}

	return c.JSON(http.StatusOK, iditem.Items[id-1])

}

// e.GET("/search", getItemFomSearching)
func getItemFomSearching(c echo.Context) error {
	//クエリパラメーターの値の取得
	keyword := c.QueryParam("keyword")
	//データベースへの接続
	db, err := sql.Open("sqlite3", DB_PATH)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}
	defer db.Close()
	//データベースから商品の検索
	//neme列にキーワードを含むものを探している
	//%は0文字以上の任意の文字列　?に"%"+keyword+"%"が入る
	rows, err := db.Query("SELECT items.name, categories.name, items.image_name FROM items JOIN categories ON items.category_id = categories.id WHERE items.name LIKE ?", "%"+keyword+"%")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}
	defer rows.Close()

	var searchResult ItemList
	for rows.Next() {
		var name, category, image string
		if err := rows.Scan(&name, &category, &image); err != nil {
			return c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
		}
		item := Item{Name: name, Category: category, Image: image}
		searchResult.Items = append(searchResult.Items, item)
	}

	return c.JSON(http.StatusOK, searchResult)

}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.DEBUG)

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


	e.GET("/items/:id", getId)

	e.GET("/search", getItemFomSearching)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
