package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const mosmixURL = "https://opendata.dwd.de/weather/local_forecasts/mos/MOSMIX_L/single_stations/10488/kml/MOSMIX_L_LATEST_10488.kmz"

// KMZ herunterladen und entpacken
func fetchKMZ() ([]byte, error) {
	resp, err := http.Get(mosmixURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	tmpFile := "mosmix.kmz"
	out, err := os.Create(tmpFile)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		return nil, err
	}

	// KMZ öffnen
	r, err := zip.OpenReader(tmpFile)
	if err != nil {
		log.Printf("error opening zip file: %v", err)
		return nil, err
	}

	defer os.Remove(tmpFile)
	defer r.Close()

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".kml") {
			rc, _ := f.Open()
			defer rc.Close()
			data, _ := io.ReadAll(rc)
			return data, nil
		}
	}

	return nil, fmt.Errorf("no kml found")
}

func main() {
	app := fiber.New()

	app.Get("/forecast", func(c *fiber.Ctx) error {
		data, err := fetchKMZ()
		if err != nil {
			log.Printf("Error while downloading: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer os.Remove("mosmix.kmz")
		log.Printf("Downloaded data length: %d", len(data))

		forecasts, err := ParseKML(data)
		if err != nil {
			log.Printf("Error while parsingn: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		if forecasts == nil {
			log.Printf("no available forecasts")
			return c.Status(404).JSON(fiber.Map{"error": "no available forecasts"})
		}

		return c.JSON(forecasts)
	})

	log.Println("Server läuft auf http://localhost:3000")
	log.Fatal(app.Listen(":3000"))
}
