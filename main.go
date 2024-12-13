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
	"github.com/joho/godotenv"
)

var IP, PORT string = "0.0.0.0", "3000"

var CORS_DOMAINS string

var TLS_ENABLED bool = true
var TLS_KEY string
var TLS_CERT string

var CAPTCHA_ENABLED bool = true
var CLOUDFLARE_TURNSTILE_SECRET_KEY string

// letterBytes is a string containing all the characters that can be used in the urlHash
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

const cloudflareVerifyUrl string = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

const maxUrlSize int = 1000
const urlHashSize int = 6

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
	if CAPTCHA_ENABLED {
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
		fmt.Println("Captcha secret key is invalid or the request failed")
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

func loadEnv() {
	godotenv.Load(".env")
	CLOUDFLARE_TURNSTILE_SECRET_KEY = os.Getenv("CLOUDFLARE_TURNSTILE_SECRET_KEY")
	CORS_DOMAINS = os.Getenv("CORS_DOMAINS")
	TLS_KEY, TLS_CERT = os.Getenv("TLS_KEY"), os.Getenv("TLS_CERT")

	if CLOUDFLARE_TURNSTILE_SECRET_KEY == "" {
		fmt.Println("CLOUDFLARE_TURNSTILE_SECRET_KEY is not set, captcha will not be enabled")
		CAPTCHA_ENABLED = false
	}

	if TLS_KEY == "" || TLS_CERT == "" {
		fmt.Println("TLS_KEY and TLS_CERT are not set, TLS will be disabled")
		TLS_ENABLED = false
	}
}

func main() {
	loadEnv()

	app := fiber.New()

	if CORS_DOMAINS != "" {
		allowedCorsDomains := strings.Split(CORS_DOMAINS, ", ")

		app.Use(cors.New(cors.Config{
			AllowOrigins: allowedCorsDomains,
		}))
	}

	app.Get("/:urlHash", getUrl)
	app.Post("/", addUrl)

	address := IP + ":" + PORT
	fmt.Println("Server listening on " + address)

	if TLS_ENABLED {
		app.Listen(address, fiber.ListenConfig{
			DisableStartupMessage: true,
			CertFile:              TLS_CERT,
			CertKeyFile:           TLS_KEY,
		})

		return
	}

	app.Listen(address, fiber.ListenConfig{
		DisableStartupMessage: true,
	})
}
