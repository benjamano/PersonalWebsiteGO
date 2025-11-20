package handlers

import (
	"PersonalWebsiteGO/config"
	"github.com/gofiber/fiber/v2"
)

func RenderLogsPage(c *fiber.Ctx) error {
	rows, err := config.DB.Query("SELECT created_at, level, message FROM log_messages WHERE created_at >= datetime('now', '-48 hours') ORDER BY created_at DESC")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving logs")
	}
	defer rows.Close()

	type LogEntry struct {
		CreatedAt string
		Level     string
		Message   string
	}

	var logs []LogEntry
	for rows.Next() {
		var log LogEntry
		if err := rows.Scan(&log.CreatedAt, &log.Level, &log.Message); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error scanning logs")
		}
		logs = append(logs, log)
	}

	return c.Render("logs/logs", fiber.Map{
		"logs": logs,
	}, "layout/base")
}