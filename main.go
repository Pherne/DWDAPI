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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}(resp.Body)

	tmpFile := "mosmix.kmz"
	out, createErr := os.Create(tmpFile)
	if createErr != nil {
		return nil, createErr
	}
	_, err = io.Copy(out, resp.Body)
	closeErr := out.Close()
	if closeErr != nil {
		return nil, closeErr
	}
	if err != nil {
		return nil, err
	}

	// KMZ öffnen
	r, err := zip.OpenReader(tmpFile)
	if err != nil {
		log.Printf("error opening zip file: %v", err)
		return nil, err
	}

	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			log.Printf("error removing file: %v", err)
		}
	}(tmpFile)
	defer func(r *zip.ReadCloser) {
		err := r.Close()
		if err != nil {
			log.Printf("error closing zip file: %v", err)
		}
	}(r)

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".kml") {
			rc, _ := f.Open()
			defer func(rc io.ReadCloser) {
				err := rc.Close()
				if err != nil {
					log.Printf("error closing file: %v", err)
				}
			}(rc)
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
		defer func() {
			err := os.Remove("mosmix.kmz")
			if err != nil {

			}
		}()
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
