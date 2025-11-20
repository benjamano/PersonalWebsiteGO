package main

import (
	"PersonalWebsiteGO/config"
	"PersonalWebsiteGO/handlers"
	"PersonalWebsiteGO/middleware"
	"PersonalWebsiteGO/background"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

func renderWithTime(c *fiber.Ctx, view string, data fiber.Map, layout string) error {
	start := time.Now()
	if data == nil {
		data = fiber.Map{}
	}
	duration := time.Since(start).Seconds()
	data["RenderTime"] = fmt.Sprintf("%.2f", duration)

	if err := c.Render(view, data, layout); err != nil {
		fmt.Println("Template render error:", err)
		return err
	}
	return nil
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize database
	if err := config.InitDatabase(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer config.CloseDatabase()

	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	engine.Reload(true)

	app.Static("/static", "./static")

	app.Get("/", func(c *fiber.Ctx) error {
		return renderWithTime(c, "index", fiber.Map{"Title": "Home"}, "layout/base")
	})

	app.Get("/projects/portfolio", func(c *fiber.Ctx) error {
		return renderWithTime(c, "projects/portfolio", fiber.Map{"Title": "Portfolio"}, "layout/base")
	})

	app.Get("/projects/lasertag", func(c *fiber.Ctx) error {
		return renderWithTime(c, "projects/lasertag", fiber.Map{"Title": "Laser Tag"}, "layout/base")
	})

	app.Get("/projects/socialmedia", func(c *fiber.Ctx) error {
		return renderWithTime(c, "projects/socialmedia", fiber.Map{"Title": "Social Media"}, "layout/base")
	})

	app.Get("/projects/websites", func(c *fiber.Ctx) error {
		return renderWithTime(c, "projects/websites", fiber.Map{"Title": "Web Development"}, "layout/base")
	})

	app.Get("/projects/software", func(c *fiber.Ctx) error {
		return renderWithTime(c, "projects/software", fiber.Map{"Title": "Software Development"}, "layout/base")
	})

	app.Get("/projects/blogs", handlers.RenderBlogsPage)

	// Authentication routes
	app.Post("/api/auth/login", handlers.Login)
	app.Get("/api/auth/check", middleware.AuthMiddleware, handlers.CheckAuth)

	app.Get("/api/blogs", handlers.GetAllBlogs)
	app.Get("/api/blogs/:id", handlers.GetBlogByID)

	// Protected API routes (require authentication)
	app.Post("/api/blogs", middleware.AuthMiddleware, handlers.CreateBlog)
	app.Put("/api/blogs/:id", middleware.AuthMiddleware, handlers.UpdateBlog)
	app.Delete("/api/blogs/:id", middleware.AuthMiddleware, handlers.DeleteBlog)

	app.Get("/api/minecraft/status", handlers.Status)
	app.Get("/api/minecraft/playerlist", handlers.PlayerList)
	app.Get("/api/minecraft/sendmessage", handlers.SendMessage)
	app.Get("/api/minecraft/getplaytime", handlers.GetPlaytime)

	app.Get("/api/proxmox/vmstatus", handlers.AllVMStatus)
	app.Get("/api/proxmox/getvmstatus", handlers.GetVMStatus)
	app.Get("/api/proxmox/getvmdetailedstatus", handlers.GetVMDetailedStatus)

	app.Get("/api/ip/currentpublicip", handlers.GetCurrentPublicIp)

	fmt.Println("Server starting on http://localhost:3000")

	background.StartPlaytimeChecker()

	fmt.Println("Background playtime checker started.")

	background.StartPublicIpValidator()

	fmt.Println("Background public IP validator started.")

	app.Listen("0.0.0.0:3000")
}
