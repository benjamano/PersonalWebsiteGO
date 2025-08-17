package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/template/html/v2"
	"time"
	"fmt"
)

func renderWithTime(c *fiber.Ctx, view string, data fiber.Map, layout string) error {
	start := time.Now()
	if data == nil {
		data = fiber.Map{}
	}
	// err := c.Render(view, data, layout)
	duration := time.Since(start).Seconds()
	data["RenderTime"] = fmt.Sprintf("%.2f", duration)
	return c.Render(view, data, layout)
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

    app.Get("/about", func(c *fiber.Ctx) error {
        return c.Render("about", fiber.Map{
            "Title": "About",
			"RenderTime": c.Locals("RenderTime"),
        }, "layout")
    })

    app.Listen(":3000")
}