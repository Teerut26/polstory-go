package services

import (
	"os/exec"
	"strconv"
	"strings"
)

func GetCoordinate(fileFullPath string) (float64, float64, error) {
	out, err := exec.Command("exiftool", "-time:all", "-location:all", "-c", `%.12f`, fileFullPath).Output()
	if err != nil {
		return 0, 0, err
	}
	outputString := strings.Split(string(out), "\n")
	var latitude float64
	var longitude float64
	for _, line := range outputString {
		key := strings.Trim(strings.Split(line, ":")[0], " ")
		if len(key) == 0 {
			continue
		}
		if key == "GPS Latitude" || key == "GPS Longitude" {
			value := strings.Trim(strings.Split(line, ":")[1], " ")
			if len(value) == 0 {
				continue
			}
			switch key {
			case "GPS Latitude":
				latitudeRaw := strings.Split(value, " ")[0]
				latitudeFloat, err := strconv.ParseFloat(latitudeRaw, 64)
				if err != nil {
					return 0, 0, err
				}
				latitude = latitudeFloat
			case "GPS Longitude":
				longitudeRaw := strings.Split(value, " ")[0]
				longitudeFloat, err := strconv.ParseFloat(longitudeRaw, 64)
				if err != nil {
					return 0, 0, err
				}
				longitude = longitudeFloat
			default:
				continue
			}
		}
	}
	return latitude, longitude, nil
}
