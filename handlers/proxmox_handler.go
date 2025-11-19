package handlers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ProxmoxClient struct {
	Host     string
	Username string
	Password string
	Ticket   string
	CSRF     string
}

func (c *ProxmoxClient) Login() error {
	url := fmt.Sprintf("https://%s:8006/api2/json/access/ticket", c.Host)
	payload := fmt.Sprintf("username=%s&password=%s", c.Username, c.Password)
	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Data struct {
			Ticket string `json:"ticket"`
			CSRF   string `json:"CSRFPreventionToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}
	c.Ticket = result.Data.Ticket
	c.CSRF = result.Data.CSRF
	return nil
}

func (c *ProxmoxClient) ListNodes() ([]string, error) {
	url := fmt.Sprintf("https://%s:8006/api2/json/nodes", c.Host)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Cookie", "PVEAuthCookie="+c.Ticket)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Data []struct {
			Node string `json:"node"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	nodes := []string{}
	for _, n := range result.Data {
		nodes = append(nodes, n.Node)
	}
	return nodes, nil
}

func (c *ProxmoxClient) ListVMStatus(node string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://%s:8006/api2/json/nodes/%s/qemu", c.Host, node)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Cookie", "PVEAuthCookie="+c.Ticket)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func GetProxMoxClient() (*ProxmoxClient, error) {
	host := os.Getenv("PROXMOX_HOST")
	username := os.Getenv("PROXMOX_USERNAME")
	password := os.Getenv("PROXMOX_PASSWORD")

	client := &ProxmoxClient{
		Host:     host,
		Username: username,
		Password: password,
	}
	return client, client.Login()
}

func AllVMStatus(c *fiber.Ctx) error {
	client, err := GetProxMoxClient()
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	nodes, err := client.ListNodes()
	if err != nil {
		fmt.Println("Failed to list nodes:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list nodes"})
	}

	statusList := make(map[string][]map[string]interface{})

	for _, node := range nodes {
		vms, err := client.ListVMStatus(node)
		if err != nil {
			fmt.Printf("Failed to list VMs for node %s: %v\n", node, err)
			continue
		}
		statusList[node] = append(statusList[node], vms...)
	}

	return c.JSON(fiber.Map{"status_list": statusList})
}

func GetVMStatus(c *fiber.Ctx) error {
	client, err := GetProxMoxClient()
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	vmId := c.Query("vmid")

	nodes, err := client.ListNodes()
	if err != nil {
		fmt.Println("Failed to list nodes:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list nodes"})
	}

	status := "VM not found"

	for _, node := range nodes {
		vms, err := client.ListVMStatus(node)
		if err != nil {
			fmt.Printf("Failed to list VMs for node %s: %v\n", node, err)
			continue
		}
		for _, vm := range vms {
			if fmt.Sprintf("%v", vm["vmid"]) == vmId {
				status = fmt.Sprintf("%v", vm["status"])
				break
			}
		}
	}

	return c.JSON(fiber.Map{"status": status})
}

func GetVMDetailedStatus(c *fiber.Ctx) error {
	client, err := GetProxMoxClient()
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	vmId := c.Query("vmid")

	nodes, err := client.ListNodes()
	if err != nil {
		fmt.Println("Failed to list nodes:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list nodes"})
	}

	for _, node := range nodes {
		vms, err := client.ListVMStatus(node)
		if err != nil {
			fmt.Printf("Failed to list VMs for node %s: %v\n", node, err)
			continue
		}
		for _, vm := range vms {
			if fmt.Sprintf("%v", vm["vmid"]) == vmId {
				return c.JSON(fiber.Map{"status": vm})
				break
			}
		}
	}

	return c.JSON(fiber.Map{"status": ""})
}
