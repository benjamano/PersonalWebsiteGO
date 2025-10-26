package handlers

import (
	"PersonalWebsiteGO/middleware"
	"os"

	"github.com/gofiber/fiber/v2"
)

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login handles user authentication
func Login(c *fiber.Ctx) error {
	var loginReq LoginRequest
	if err := c.BodyParser(&loginReq); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check credentials against environment variables
	adminUsername := os.Getenv("ADMIN_USERNAME")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if loginReq.Username != adminUsername || loginReq.Password != adminPassword {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid username or password",
		})
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(loginReq.Username)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.JSON(fiber.Map{
		"token":   token,
		"message": "Login successful",
	})
}

// CheckAuth verifies if the user is authenticated
func CheckAuth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"authenticated": true,
		"message":       "User is authenticated",
	})
}
