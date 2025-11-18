package handlers

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/james4k/rcon"
)

func GetRCONClient() *rcon.RemoteConsole {
	rconClient, err := rcon.Dial(os.Getenv("MINECRAFT_RCON_ADDRESS"), os.Getenv("MINECRAFT_RCON_PASSWORD"))
	if err != nil {
		fmt.Println("Error connecting to RCON:", err)
		return nil
	}

	return rconClient
}

func Status(c *fiber.Ctx) error {
	rconClient := GetRCONClient()
	if rconClient == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to connect to RCON"})
	}
	defer rconClient.Close()

	start := time.Now()
	_, err := rconClient.Write("list")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	listResponse, _, err := rconClient.Read()
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	_, err = rconClient.Write("version")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var playersOnline, maxPlayers int
	fmt.Sscanf(listResponse, "There are %d of a max of %d players online", &playersOnline, &maxPlayers)

	return c.JSON(fiber.Map{
		"online":         true,
		"latency":        latency,
		"players_online": playersOnline,
		"max_players":    maxPlayers,
	})
}

func PlayerList(c *fiber.Ctx) error {
	rconClient := GetRCONClient()
	if rconClient == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to connect to RCON"})
	}
	defer rconClient.Close()

	_, err := rconClient.Write("list")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	listResponse, _, err := rconClient.Read()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	listResponse = strings.Split(listResponse, ": ")[1]

	playerList := strings.Split(strings.TrimSpace(listResponse), "\n")

	return c.JSON(fiber.Map{
		"players": playerList,
	})
}

func SendMessage(c *fiber.Ctx) error {
	rconClient := GetRCONClient()
	if rconClient == nil {
		return fmt.Errorf("failed to connect to RCON")
	}
	defer rconClient.Close()

	message := c.FormValue("message")
	if message == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Message is required"})
	}

	_, err := rconClient.Write(fmt.Sprintf("say %s", message))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}