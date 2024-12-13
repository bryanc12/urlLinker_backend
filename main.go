package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// letterBytes is a string containing all the characters that can be used in the urlHash
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

const cloudflareVerifyUrl string = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

const maxUrlSize int = 1000
const urlHashSize int = 6

var CORS_DOMAINS string = os.Getenv("CORS_DOMAINS")

var TLS_ENABLED bool = os.Getenv("TLS_ENABLED") == "true"
var TLS_SECRET string = os.Getenv("TLS_PRIVATE")
var TLS_CERT string = os.Getenv("TLS_CERT")

var CLOUDFLARE_TURNSTILE_SECRET_KEY string = os.Getenv("CLOUDFLARE_TURNSTILE_SECRET_KEY")
var captchaEnabled bool = CLOUDFLARE_TURNSTILE_SECRET_KEY != ""

var urlCount int = 0
var urlMap map[string]string = make(map[string]string)

func getUrl(c fiber.Ctx) error {
	if urlCount == 0 {
		return c.SendStatus(fiber.StatusNotFound)
	}

	urlHash := c.Params("urlHash")
	url, ok := urlMap[urlHash]

	if !ok {
		return c.SendStatus(fiber.StatusNotFound)
	}

	c.SendString(url)
	return nil
}

func addUrl(c fiber.Ctx) error {
	if captchaEnabled {
		if !verifyCaptcha(c) {
			return c.SendStatus(fiber.StatusBadRequest)
		}
	}

	requestUrl := c.Query("url")
	if requestUrl == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if urlCount == maxUrlSize {
		urlMap = make(map[string]string)
		urlCount = 0
	}

	urlParsed, err := url.ParseRequestURI(requestUrl)
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// Random string with the length of urlHashSize
	urlHash := make([]byte, urlHashSize)
	for i := range urlHash {
		urlHash[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	urlHashString := string(urlHash)

	urlMap[urlHashString] = urlParsed.String()
	urlCount++

	c.SendString(urlHashString)
	return nil
}

func verifyCaptcha(c fiber.Ctx) bool {
	// Check if the captcha token is present
	captchaToken := c.Query("captcha_token")
	if captchaToken == "" {
		return false
	}

	// Check if the CF-Connecting-IP header is present and is a valid IP
	ip := c.Request().Header.Peek("CF-Connecting-IP")
	ipString := string(ip)
	if net.ParseIP(ipString) == nil {
		return false
	}

	params := url.Values{}
	params.Add("secret", CLOUDFLARE_TURNSTILE_SECRET_KEY)
	params.Add("response", captchaToken)
	params.Add("remoteip", ipString)

	payload := bytes.NewBufferString(params.Encode())

	request, err := http.NewRequest("POST", cloudflareVerifyUrl, payload)
	if err != nil {
		return false
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := http.DefaultClient.Do(request)
	if err != nil || response.StatusCode != http.StatusOK {
		return false
	}

	type CaptchaCheckResponse struct {
		Success bool `json:"success"`
	}

	var captchaCheckResponse CaptchaCheckResponse
	err = json.NewDecoder(response.Body).Decode(&captchaCheckResponse)
	if err != nil {
		return false
	}

	return captchaCheckResponse.Success
}

func main() {
	ip, port := os.Getenv("IP"), os.Getenv("PORT")

	if CLOUDFLARE_TURNSTILE_SECRET_KEY == "" {
		fmt.Println("CLOUDFLARE_TURNSTILE_SECRET_KEY is not set, captcha will not be enabled")
		captchaEnabled = false
	}

	if ip == "" {
		fmt.Println("IP is not set, defaulting to 0.0.0.0")
		ip = "0.0.0.0"
	}

	if port == "" {
		fmt.Println("PORT is not set, defaulting to 9999")
		port = "9999"
	}

	app := fiber.New()

	if CORS_DOMAINS != "" {
		allowedCorsDomains := strings.Split(CORS_DOMAINS, ", ")

		app.Use(cors.New(cors.Config{
			AllowOrigins: allowedCorsDomains,
		}))
	}

	app.Get("/:urlHash", getUrl)
	app.Post("/", addUrl)

	app.Listen(ip + ":" + port)
}
