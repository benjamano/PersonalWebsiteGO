package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func RenderServerStatusPage(c *fiber.Ctx) error {
	return c.Render("servicestatus/servicestatus", fiber.Map{
		"Title": "Service Status",
	}, "layout/base")
}