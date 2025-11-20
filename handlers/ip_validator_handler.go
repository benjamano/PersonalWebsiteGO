package handlers

import (
	"fmt"
	"net/http"
	"encoding/json"
	"os"
	"github.com/gofiber/fiber/v2"
	"PersonalWebsiteGO/config"
	"PersonalWebsiteGO/models"
	"bytes"
)

func GetCurrentPublicIp(c *fiber.Ctx) error {
	ip, err := _GetCurrentPublicIp()

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"ip": ip})
}

func _GetCurrentPublicIp() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=text")

	if err != nil {
		return "", fmt.Errorf("failed to fetch public IP: %w", err)
	}

	defer resp.Body.Close()

	var ip string
	_, err = fmt.Fscan(resp.Body, &ip)
	if err != nil {
		return "", fmt.Errorf("failed to read public IP: %w", err)
	}

	return ip, nil
}

func ValidatePublicIpWithCloudflare(ip string) (bool, error) {
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")

	if apiToken == "" || zoneID == "" {
		return false, fmt.Errorf("missing Cloudflare API credentials")
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?type=A", zoneID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer " + apiToken)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to fetch DNS records: %w", err)
	}
	defer resp.Body.Close()
	
	var result struct {
		Success bool `json:"success"`
		Result  []struct {
			Content string `json:"content"`
			Name    string `json:"name"`
		} `json:"result"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode cloudflare response: %w", err)
	}
	
	if !result.Success {
		return false, fmt.Errorf("cloudflare api request failed")
	}
	
	for _, record := range result.Result {
		if record.Content != ip {
			config.LogMessage("ERROR", fmt.Sprintf("DNS record %s has IP %s, expected %s", record.Name, record.Content, ip))
			return false, nil
		}
	}
	
	return true, nil
}

func UpdatePublicIpOnCloudflare(ip string) error {
	var lastIpRecord models.PublicIpUpdate

	rows, err := config.DB.Query("SELECT new_public_ip_address FROM public_ip_updates ORDER BY changed_at DESC LIMIT 1")
	if err != nil {
		config.LogMessage("ERROR", "Error fetching last public IP record: "+err.Error())
		return err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&lastIpRecord.NewPublicIpAddress); err != nil {
			config.LogMessage("ERROR", "Error scanning last public IP record: "+err.Error())
			return err
		}
	}

	_, err = config.DB.Exec("INSERT INTO public_ip_updates (new_public_ip_address, old_public_ip_address) VALUES (?, ?)", ip, lastIpRecord.NewPublicIpAddress)
	if err != nil {
		config.LogMessage("ERROR", "Error inserting new public IP record: "+err.Error())
	}

	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")

	if apiToken == "" || zoneID == "" {
		return fmt.Errorf("missing Cloudflare API credentials")
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?type=A", zoneID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch DNS records: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
		Result  []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Content string `json:"content"`
			Proxied bool   `json:"proxied"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode Cloudflare response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("cloudflare api request failed")
	}

	for _, record := range result.Result {
		if record.Content != ip {
			updateURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneID, record.ID)
			body := map[string]interface{}{
				"type":    "A",
				"name":    record.Name,
				"content": ip,
				"ttl":     1,
				"proxied": record.Proxied,
			}
			jsonBody, _ := json.Marshal(body)

			updateReq, err := http.NewRequest("PUT", updateURL, bytes.NewBuffer(jsonBody))
			if err != nil {
				config.LogMessage("ERROR", fmt.Sprintf("Failed to create update request for %s: %v", record.Name, err))
				continue
			}
			updateReq.Header.Set("Authorization", "Bearer "+apiToken)
			updateReq.Header.Set("Content-Type", "application/json")

			updateResp, err := client.Do(updateReq)
			if err != nil {
				config.LogMessage("ERROR", fmt.Sprintf("Failed to update DNS record %s: %v", record.Name, err))
				continue
			}
			updateResp.Body.Close()
			config.LogMessage("INFO", fmt.Sprintf("Updated DNS record %s to IP %s", record.Name, ip))
		}
	}

	return nil
}

func CheckToUpdatePublicIp() {
	ip, err := _GetCurrentPublicIp()
	if err != nil {
		config.LogMessage("ERROR", "Error fetching current public IP: "+err.Error())
		return
	}

	isValid, err := ValidatePublicIpWithCloudflare(ip)
	if err != nil {
		config.LogMessage("ERROR", "Error validating public ip with cloudflare: "+err.Error())
		return
	}

	if !isValid {
		config.LogMessage("INFO", "Public IP does not match Cloudflare's record, updating...")

		err := UpdatePublicIpOnCloudflare(ip)
		if err != nil {
			config.LogMessage("ERROR", "Error updating public IP on Cloudflare: "+err.Error())
			return
		}

		config.LogMessage("INFO", "Public IP updated successfully.")
	} else {
		// config.LogMessage("INFO", "Public IP is valid and matches Cloudflare's record.")
	}
}