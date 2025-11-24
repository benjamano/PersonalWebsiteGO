package handlers

import (
	"PersonalWebsiteGO/config"
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
		config.LogMessage("ERROR", "Error connecting to RCON: "+err.Error())
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
	playerList, err := GetPlayerList()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

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

func GetPlayerList() ([]string, error) {
	rconClient := GetRCONClient()
	if rconClient == nil {
		return nil, fmt.Errorf("failed to connect to RCON")
	}
	defer rconClient.Close()

	_, err := rconClient.Write("list")
	if err != nil {
		return nil, fmt.Errorf("failed to send RCON command: %w", err)
	}

	listResponse, _, err := rconClient.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read RCON response: %w", err)
	}

	parts := strings.SplitN(listResponse, ": ", 2)
	if len(parts) < 2 {
		return []string{}, nil
	}

	afterColon := strings.TrimSpace(parts[1])
	if afterColon == "" {
		return []string{}, nil
	}

	rawPlayers := strings.FieldsFunc(afterColon, func(r rune) bool {
		return r == '\n' || r == ',' 
	})

	players := make([]string, 0, len(rawPlayers))
	for _, p := range rawPlayers {
		p = strings.TrimSpace(p)
		if p != "" {
			players = append(players, p)
		}
	}

	return players, nil
}

func CheckAndUpdatePlaytime() {
	playerList, err := GetPlayerList()
	if err != nil {
		config.LogMessage("ERROR", "Error getting player list: "+err.Error())
		return
	}
	for _, player := range playerList {
		config.LogMessage("INFO", fmt.Sprintf("Updating playtime for player: %s", player))

		rows, err := config.DB.Query("SELECT playtime, id FROM user_playtime WHERE user_name = ? AND date = ?", player, time.Now().Format("02-01-2006"))
		if err != nil {
			config.LogMessage("ERROR", "Database query error: "+err.Error())
			continue
		}

		var playtimeMinutes int
		var id int
		if rows.Next() {
			if err := rows.Scan(&playtimeMinutes, &id); err != nil {
				config.LogMessage("ERROR", "Row scan error: "+err.Error())
				rows.Close()
				continue
			}
			rows.Close()
			playtimeMinutes += 10
			_, err = config.DB.Exec("UPDATE user_playtime SET playtime = ?, last_login = ? WHERE id = ?",
				playtimeMinutes, time.Now(), id)
			if err != nil {
				config.LogMessage("ERROR", "Database update error: "+err.Error())
				continue
			}
			config.LogMessage("INFO", fmt.Sprintf("Updated playtime for player %s: %d minutes", player, playtimeMinutes))
		} else {
			rows.Close()
			playtimeMinutes = 10
			_, err = config.DB.Exec("INSERT INTO user_playtime (user_name, playtime, date, last_login) VALUES (?, ?, ?, ?)",
				player, playtimeMinutes, time.Now().Format("02-01-2006"), time.Now())
			if err != nil {
				config.LogMessage("ERROR", "Database insert error: "+err.Error())
				continue
			}
			config.LogMessage("INFO", fmt.Sprintf("Inserted new playtime for player %s: %d minutes", player, playtimeMinutes))
		}
	}
}

func GetPlaytime(c *fiber.Ctx) error {
	date := c.Query("date", time.Now().Format("02-01-2006"))

	rows, err := config.DB.Query("SELECT user_name, playtime FROM user_playtime WHERE date = ?", date)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	defer rows.Close()
	playtimeData := make(map[string]int)

	for rows.Next() {
		var userName string
		var playtime int
		if err := rows.Scan(&userName, &playtime); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		playtimeData[userName] = playtime
	}

	return c.JSON(fiber.Map{"playtime": playtimeData})
}
