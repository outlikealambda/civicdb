package wmaddress

import (
	"math"
	"testing"
)

func TestCalculateDistance(t *testing.T) {
	honoluluLat := 21.296834
	honoluluLon := -157.856655

	torontoLat := 43.649753
	torontoLon := -79.374704

	expectedDistance := float64(7499065)
	distance, _ := CalculateDistance(honoluluLat, honoluluLon, torontoLat, torontoLon)

	if distanceError := math.Abs(distance - expectedDistance); distanceError > 10 {
		t.Errorf("Calculated distance in meters should have been: %v, but was: %v, off by ~%vmeters", expectedDistance, distance, int(distance-expectedDistance))
	}
}

func TestExtractCoordinates(t *testing.T) {
	coordString := "(123, 345)"

	lat, lon, _ := ExtractCoordinates(coordString)
	if lat != 123 || lon != 345 {
		t.Errorf("Expected { 123, 345 }, but got: %v, %v", lat, lon)
	}
}

func TestCalculateApproximateDistance(t *testing.T) {
	d := CalculateApproximateDistance(6, 0, 6, 60)

	if math.Abs(d-6) > 0.01 {
		t.Errorf("expected a hypoteneuse of 6, but got: %v", d)
	}
}

func TestCalculateApproximateDistanceWrappedHeadings(t *testing.T) {
	d := CalculateApproximateDistance(6, 330, 6, 30)

	if math.Abs(d-6) > 0.01 {
		t.Errorf("expected a hypoteneuse of 6, but got: %v", d)
	}
}
