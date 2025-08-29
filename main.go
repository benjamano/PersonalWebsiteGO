package main

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/reverseproxy"
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
	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

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

	app.Get("/projects/software", func(c *fiber.Ctx) error {
		return renderWithTime(c, "projects/software", fiber.Map{"Title": "Software Development"}, "layout/base")
	})

	octoprintProxy := reverseproxy.NewReverseProxy("192.168.1.87")

	app.All("/octoprint/*", func(c *fiber.Ctx) error {
		// Strip the "/octoprint" prefix before forwarding
		newPath := c.OriginalURL()[len("/octoprint"):]
		c.Path(newPath)
		octoprintProxy.ServeHTTP(c.Context())
		return nil
	})

	fmt.Println("Server starting on http://localhost:3000")

	app.Listen("0.0.0.0:3000")
}
