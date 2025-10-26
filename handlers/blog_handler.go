package handlers

import (
	"PersonalWebsiteGO/config"
	"PersonalWebsiteGO/models"
	"fmt"
	"html/template"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetAllBlogs retrieves all blog posts
func GetAllBlogs(c *fiber.Ctx) error {
	rows, err := config.DB.Query("SELECT id, title, content, author, created_at, updated_at FROM blogs ORDER BY created_at DESC")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var blogs []models.Blog
	for rows.Next() {
		var blog models.Blog
		err := rows.Scan(&blog.ID, &blog.Title, &blog.Content, &blog.Author, &blog.CreatedAt, &blog.UpdatedAt)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		blogs = append(blogs, blog)
	}

	return c.JSON(blogs)
}

// GetBlogByID retrieves a single blog post by ID
func GetBlogByID(c *fiber.Ctx) error {
	id := c.Params("id")

	var blog models.Blog
	err := config.DB.QueryRow("SELECT id, title, content, author, created_at, updated_at FROM blogs WHERE id = ?", id).
		Scan(&blog.ID, &blog.Title, &blog.Content, &blog.Author, &blog.CreatedAt, &blog.UpdatedAt)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Blog not found"})
	}

	return c.JSON(blog)
}

// CreateBlog creates a new blog post
func CreateBlog(c *fiber.Ctx) error {
	var blog models.Blog
	if err := c.BodyParser(&blog); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	now := time.Now()
	result, err := config.DB.Exec(
		"INSERT INTO blogs (title, content, author, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		blog.Title, blog.Content, blog.Author, now, now,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	id, _ := result.LastInsertId()
	blog.ID = int(id)
	blog.CreatedAt = now
	blog.UpdatedAt = now

	return c.Status(201).JSON(blog)
}

// UpdateBlog updates an existing blog post
func UpdateBlog(c *fiber.Ctx) error {
	id := c.Params("id")

	var blog models.Blog
	if err := c.BodyParser(&blog); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	now := time.Now()
	result, err := config.DB.Exec(
		"UPDATE blogs SET title = ?, content = ?, author = ?, updated_at = ? WHERE id = ?",
		blog.Title, blog.Content, blog.Author, now, id,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Blog not found"})
	}

	return c.JSON(fiber.Map{"message": "Blog updated successfully"})
}

// DeleteBlog deletes a blog post
func DeleteBlog(c *fiber.Ctx) error {
	id := c.Params("id")

	result, err := config.DB.Exec("DELETE FROM blogs WHERE id = ?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Blog not found"})
	}

	return c.JSON(fiber.Map{"message": "Blog deleted successfully"})
}

// RenderBlogsPage renders the blogs page with all blog posts
func RenderBlogsPage(c *fiber.Ctx) error {
	rows, err := config.DB.Query("SELECT id, title, content, author, created_at, updated_at FROM blogs ORDER BY created_at DESC")
	if err != nil {
		fmt.Println("Error fetching blogs:", err)
		return c.Render("projects/blogs", fiber.Map{
			"Title": "Blogs",
			"Blogs": []models.Blog{},
			"Error": "Failed to load blogs",
		}, "layout/base")
	}
	defer rows.Close()

	var blogs []models.Blog
	for rows.Next() {
		var blog models.Blog
		err := rows.Scan(&blog.ID, &blog.Title, &blog.Content, &blog.Author, &blog.CreatedAt, &blog.UpdatedAt)
		if err != nil {
			fmt.Println("Error scanning blog:", err)
			continue
		}
		// Convert content to template.HTML for safe rendering
		blog.Content = template.HTML(blog.Content)
		blogs = append(blogs, blog)
	}

	return c.Render("projects/blogs", fiber.Map{
		"Title": "Blogs",
		"Blogs": blogs,
	}, "layout/base")
}
