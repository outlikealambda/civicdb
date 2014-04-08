package wmaddress

import (
	"errors"
	"github.com/StefanSchroeder/Golang-Ellipsoid/ellipsoid"
	"math"
	"strconv"
	"strings"
)

// does this function as a static variable?
var geo1 = ellipsoid.Init("WGS84", ellipsoid.Degrees, ellipsoid.Meter, ellipsoid.Longitude_is_symmetric, ellipsoid.Bearing_is_symmetric)

// use the law of cosines to approximate distance -- cheaper than recalculating with the ellipsoid
func CalculateApproximateDistance(distanceA, bearingA, distanceB, bearingB float64) float64 {

	// find the convex angle based on compass bearings
	convexAngle := math.Abs(bearingA - bearingB)
	if convexAngle == 180 {
		// law of cosines will fail.  unlikely, but still...
		return 1000000000
	}

	if convexAngle > 180 {
		convexAngle = 360 - convexAngle
	}

	convexAngle = convexAngle / 180 * math.Pi

	// c^2 = a^2 + b^2 - 2ab*cos(C)
	d := math.Sqrt(distanceA*distanceA + distanceB*distanceB - 2*distanceA*distanceB*math.Cos(convexAngle))

	return d
}

func CalculateDistance(latA, lonA, latB, lonB float64) (distance, bearing float64) {
	distance, bearing = geo1.To(latA, lonA, latB, lonB)

	return
}

func ExtractCoordinates(coordString string) (lat float64, lon float64, err error) {
	// coordinates = Coordinates{}

	if strings.Index(coordString, "(") < 0 || strings.Index(coordString, ")") < 0 {
		err = errors.New("improperly formatted coordinates")
		return
	}

	commaIndex := strings.Index(coordString, ",")

	latAsString := coordString[1:commaIndex]
	lonAsString := coordString[commaIndex+2 : len(coordString)-1]

	var latErr error
	var longErr error

	lat, latErr = strconv.ParseFloat(latAsString, 64)
	lon, longErr = strconv.ParseFloat(lonAsString, 64)

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
