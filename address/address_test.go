package address

import (
	"math"
	"testing"
)

func TestCalculateDistance(t *testing.T) {
	honoluluOffice := new(Coordinates)
	torontoOffice := new(Coordinates)

	honoluluOffice.latitude = 21.296834
	honoluluOffice.longitude = -157.856655

	torontoOffice.latitude = 43.649753
	torontoOffice.longitude = -79.374704

	expectedDistance := float64(7499065)
	distance := CalculateDistance(*honoluluOffice, *torontoOffice)

	if distanceError := math.Abs(distance - expectedDistance); distanceError > 10 {
		t.Errorf("Calculated distance in meters should have been: %v, but was: %v, off by ~%vmeters", expectedDistance, distance, int(distance-expectedDistance))
	}
}

func TestExtractCoordinates(t *testing.T) {
	coordString := "(123, 345)"

	ExtractCoordinates(coordString)
}
