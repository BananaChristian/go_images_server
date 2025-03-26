package main

import (
	"database/sql"
	"encoding/base64"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

var db *sql.DB

type Image struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Image string `json:"image"`
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "images.db")
	if err != nil {
		log.Fatalf("Failed to connect to database")
		return
	}

	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS images(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		image LONGTEXT NOT NULL
		)`)

	if err != nil {
		log.Fatalf("Failed to create table")
		return
	}
}

func handleImageUpload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 5<<20)
	var image Image
	if err := c.ShouldBindJSON(&image); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data", "details": err.Error()})
		return
	}

	if _, err := base64.StdEncoding.DecodeString(image.Image); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image data", "details": err.Error()})
		return
	}

	_, err := db.Exec(
		"INSERT INTO images VALUES(NULL,?,?)", image.Title, image.Image)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert image", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image sucessfully updated"})

}

func deleteImage(c *gin.Context) {

	title := c.Param("title")

	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data title is required"})
		return
	}

	result, err := db.Exec(
		"DELETE FROM images WHERE title=?", title)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image", "details": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image successfully deleted", "title": title})
}

func getAllImages(c *gin.Context) {
	rows, err := db.Query(
		"SELECT id, title, image FROM images ORDER BY id LIMIT 50")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get image", "details": err.Error()})
		return
	}

	defer rows.Close()

	var images []Image
	for rows.Next() {
		var image Image
		if err := rows.Scan(&image.ID, &image.Title, &image.Image); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan image data"})
			return
		}

		images = append(images, image)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occured while iterating over the rows", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Images successfully retrieved", "images": images})

}

func getImageByTitle(c *gin.Context) {
	title := c.Param("title")
	rows, err := db.Query(
		"SELECT id, title, image FROM IMAGES WHERE title=? ORDER BY id LIMIT 50", title)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get image by title", "details": err.Error()})
		return
	}

	defer rows.Close()

	var images []Image
	for rows.Next() {
		var image Image
		if err := rows.Scan(&image.ID, &image.Title, &image.Image); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan image data", "details": err.Error()})
			return
		}

		images = append(images, image)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sucessfully retrieved images by title", "images": images})
}

func main() {
	initDB()
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Origin"},
		AllowCredentials: true,
	}))

	r.POST("/upload", handleImageUpload)
	r.GET("/getAllImages", getAllImages)
	r.GET("/getImageByTitle/:title", getImageByTitle)
	r.DELETE("/delete/:title", deleteImage)
	r.Run()

}
