package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

type User struct {
	ID       uint   `json:"ID" gorm:"primaryKey"`
	Name     string `json:"Name"`
	Birthday string `json:"Birthday"`
}

// Load environment variables
func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
}

// Initialize database connection
func initDB() {
	var err error
	dbType := os.Getenv("DB_TYPE")

	switch dbType {
	case "postgres":
		dsn := os.Getenv("DATABASE_URL")
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "sqlite":
		dsn := "users.db"
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	default:
		log.Fatal("Unsupported database type. Set DB_TYPE to 'postgres' or 'sqlite'")
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.AutoMigrate(&User{})
	log.Println("Database connected and migrated successfully.")
}

// Fetch all users
func getUsers(c echo.Context) error {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch users"})
	}
	return c.JSON(http.StatusOK, users)
}

// Fetch a  user
func getUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}
	var user User
	if err := db.First(&user, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
	}
	return c.JSON(http.StatusOK, user)
}

// Create a new user
func createUser(c echo.Context) error {
	user := new(User)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	if user.Name == "" || user.Birthday == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name and Birthday are required"})
	}

	if err := db.Create(user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
	}
	return c.JSON(http.StatusCreated, user)
}

// Update an existing user
func updateUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	var user User
	if err := db.First(&user, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	updatedUser := new(User)
	if err := c.Bind(updatedUser); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Update user fields if provided
	if updatedUser.Name != "" {
		user.Name = updatedUser.Name
	}
	if updatedUser.Birthday != "" {
		user.Birthday = updatedUser.Birthday
	}

	if err := db.Save(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user"})
	}

	return c.JSON(http.StatusOK, user)
}

// Delete a user
func deleteUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	var user User
	if err := db.First(&user, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	if err := db.Delete(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

func main() {
	loadEnv()
	initDB()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/users", getUsers)
	e.GET("/users/:id", getUser)
	e.POST("/users", createUser)
	e.PUT("/users/:id", updateUser)
	e.DELETE("/users/:id", deleteUser)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
