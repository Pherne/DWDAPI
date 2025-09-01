package main

import "encoding/xml"

type KML struct {
	XMLName  xml.Name `xml:"kml"`
	Document Document `xml:"Document"`
}

type Document struct {
	ExtendedData ProductDefinition `xml:"ExtendedData>ProductDefinition"`
	Placemark    Placemark         `xml:"Placemark"`
}

type ProductDefinition struct {
	ForecastTimeSteps TimeSteps `xml:"ForecastTimeSteps"`
}

type TimeSteps struct {
	TimeStep []string `xml:"TimeStep"`
}

type Placemark struct {
	Name         string       `xml:"name"`
	Description  string       `xml:"description"`
	ExtendedData ForecastData `xml:"ExtendedData"`
	Point        Point        `xml:"Point"`
}

type ForecastData struct {
	Forecast []Forecast `xml:"Forecast"`
}

type Forecast struct {
	ElementName string `xml:"elementName,attr"`
	Value       string `xml:"value"`
}

type Point struct {
	Coordinates string `xml:"coordinates"`
}

type WeatherData struct {
	StationName string            `json:"stationName"`
	Forecasts   []ForecastElement `json:"forecasts"`
}

type ForecastElement struct {
	Timestamp     string   `json:"timestamp"`
	Temp          float64  `json:"temperatureCelsius"`
	MinTemp       *float64 `json:"minTemperatureCelsius,omitempty"`
	MaxTemp       *float64 `json:"maxTemperatureCelsius,omitempty"`
	Pressure      float64  `json:"pressure"`
	WindDirection string   `json:"windDirection"`
	WindSpeed     float64  `json:"windSpeed"`
	PreciChance   *int     `json:"precipitationChance,omitempty"`
	PreciAmount   float64  `json:"precipitationAmount"`
	SnowChance    float64  `json:"snowChance"`
	FogChance     float64  `json:"fogChance"`
}
