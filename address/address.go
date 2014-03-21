package address

import (
	"errors"
	"github.com/StefanSchroeder/Golang-Ellipsoid/ellipsoid"
	"strconv"
	"strings"
)

type Coordinates struct {
	longitude float64
	latitude  float64
}

func New(latitude, longitude float64) Coordinates {
	return Coordinates{latitude, longitude}
}

func NewEmpty() Coordinates {
	return Coordinates{}
}

func (coords *Coordinates) Longitude() float64 {
	return coords.longitude
}

func (coords *Coordinates) Latitude() float64 {
	return coords.latitude
}

func (coords *Coordinates) SetLongitude(longitude float64) {
	coords.longitude = longitude
}

func (coords *Coordinates) SetLatitude(latitude float64) {
	coords.latitude = latitude
}

func CalculateDistance(address, reference Coordinates) float64 {

	geo1 := ellipsoid.Init("WGS84", ellipsoid.Degrees, ellipsoid.Meter, ellipsoid.Longitude_is_symmetric, ellipsoid.Bearing_is_symmetric)

	distance, _ := geo1.To(address.latitude, address.longitude, reference.latitude, reference.longitude)

	return distance
}

func ExtractCoordinates(coordString string) (coordinates Coordinates, err error) {
	// coordinates = Coordinates{}

	if strings.Index(coordString, "(") < 0 || strings.Index(coordString, ")") < 0 {
		err = errors.New("improperly formatted coordinates")
		return
	}

	commaIndex := strings.Index(coordString, ",")

	latitude := coordString[1:commaIndex]
	longitude := coordString[commaIndex+2 : len(coordString)-1]

	var latErr error
	var longErr error

	coordinates.latitude, latErr = strconv.ParseFloat(latitude, 64)
	coordinates.longitude, longErr = strconv.ParseFloat(longitude, 64)

	if latErr != nil {
		err = latErr
		return
	}

	if longErr != nil {
		err = longErr
		return
	}

	return
}
