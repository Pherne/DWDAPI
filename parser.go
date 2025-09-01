package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"golang.org/x/net/html/charset"
	"log"
	"strings"
)

func ParseKML(data []byte) (*WeatherData, error) {
	var kml KML

	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.CharsetReader = charset.NewReaderLabel

	if err := decoder.Decode(&kml); err != nil {
		return nil, fmt.Errorf("error decoding KML: %v", err)
	}

	// Create a weather data structure
	weatherData := &WeatherData{
		StationName: kml.Document.Placemark.Description,
		Forecasts:   make([]ForecastElement, 0),
	}

	timeSteps := kml.Document.ExtendedData.ForecastTimeSteps.TimeStep

	// Create a map to store temporary forecast data
	forecastMap := make(map[string]*ForecastElement)

	// Initialize forecast elements for each timestamp
	for _, timeStep := range timeSteps {
		forecastMap[timeStep] = &ForecastElement{
			Timestamp: timeStep,
		}
	}

	// Process each forecast element
	for _, forecast := range kml.Document.Placemark.ExtendedData.Forecast {
		values := strings.Fields(forecast.Value)

		for i, value := range values {
			if i >= len(timeSteps) {
				break
			}

			timeStep := timeSteps[i]
			elem, exists := forecastMap[timeStep]
			if !exists {
				continue
			}

			var val float64
			_, scanErr := fmt.Sscanf(value, "%f", &val)
			if scanErr != nil {
				if value == "-" {
					continue
				}
				log.Printf("Error parsing %s: %v in %s", value, scanErr, forecast.ElementName)
				continue
			}

			// explaination for each parameter: https://dwd-geoportal.de/products/G_FJM/
			switch forecast.ElementName {
			case "TTT":
				elem.Temp = val - 273.15
			case "TN":
				if value != "-" {
					minTempC := val - 273.15
					elem.MinTemp = &minTempC
				}
			case "TX":
				if value != "-" {
					maxTempC := val - 273.15
					elem.MaxTemp = &maxTempC
				}
			case "PPPP":
				elem.Pressure = val

			case "DD":
				if val > 340 || val <= 20 {
					elem.WindDirection = "N"
				} else if val > 20 && val <= 70 {
					elem.WindDirection = "NE"
				} else if val > 70 && val <= 110 {
					elem.WindDirection = "E"
				} else if val > 110 && val <= 150 {
					elem.WindDirection = "SE"
				} else if val > 150 && val <= 200 {
					elem.WindDirection = "S"
				} else if val > 200 && val <= 240 {
					elem.WindDirection = "SW"
				} else if val > 240 && val <= 280 {
					elem.WindDirection = "W"
				} else {
					elem.WindDirection = "NW"
				}
			case "FF":
				elem.WindSpeed = val
			case "Rh00":
				if value != "-" {
					pChance := int(val)
					elem.Precipitation = &pChance
				}
			case "RR1c":
				// unit is mm
				elem.PAmount = val
			case "RRS1c":
				elem.SnowChance = val
			case "wwM":
				elem.FogChance = val
			case "wwS":
				elem.HailChance = val
			case "wwT":
				elem.ThunderChance = val
			case "wwF":
				elem.FreezingRainChance = val

			}

		}
	}

	// Convert map to slice
	for _, timeStep := range timeSteps {
		if elem, exists := forecastMap[timeStep]; exists {
			weatherData.Forecasts = append(weatherData.Forecasts, *elem)
		}
	}

	if len(weatherData.Forecasts) == 0 {
		return nil, fmt.Errorf("no forecast data found")
	}

	return weatherData, nil
}
