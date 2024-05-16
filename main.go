package main

import (
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Endpoint struct {
	Path      string
	TargetURL string
}

var endpoints = []Endpoint{
	{Path: "/service", TargetURL: "http://localhost:3005/CGaaS-Angular-Ui-Generator/api"},
}

func proxyRequest(c *fiber.Ctx, targetURL string) error {
	client := resty.New()

	// Convert headers to the expected type
	headers := make(map[string]string)
	for key, values := range c.GetReqHeaders() {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Convert fasthttp.Args to url.Values
	queryParams := make(url.Values)
	c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
		queryParams.Add(string(key), string(value))
	})

	// Parse the target URL
	// parsedURL, err := url.Parse(targetURL)
	// if err != nil {
	//     return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	// }

	// Join base URL and path without the "/service" part
	//fullURL := strings.TrimSuffix(parsedURL.String(), "/service") + string(c.Request().RequestURI())

	// Forward the request to the target backend service
	resp, err := client.R().
		SetQueryParamsFromValues(queryParams).
		SetHeaders(headers).
		SetBody(c.Body()).
		Execute(c.Method(), targetURL)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.Status(resp.StatusCode()).Send(resp.Body())
}

func main() {
	app := fiber.New()

	// Enable logger middleware
	app.Use(logger.New())

	// Register each endpoint
	for _, endpoint := range endpoints {
		path := endpoint.Path
		targetURL := endpoint.TargetURL

		// Register route for all HTTP methods
		app.All(path, func(c *fiber.Ctx) error {
			return proxyRequest(c, targetURL)
		})
	}

	// Start the server
	app.Listen(":3000")
}
