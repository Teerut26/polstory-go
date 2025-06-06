package route

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/freetype/truetype"
	"github.com/teerut26/polstory-go/services"
)

func Generate916Handler(c *fiber.Ctx) error {
	imagePayload, err := c.FormFile("image")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	rotateAngle := 0.0
	rotateAngle, err = strconv.ParseFloat(c.FormValue("rotateAngle"), 64)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if rotateAngle > 360.0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Rotate angle must be less than 360.0",
		})
	}

	scale := 1.0
	scale, err = strconv.ParseFloat(c.FormValue("scale"), 64)
	if scale > 3.0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Scale must be less than 3.0",
		})
	} else if scale < 0.5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Scale must be greater than 0.5",
		})
	}

	var (
		canvasWidth  = 1133.0 * scale
		canvasHeight = 2016.0 * scale
		zoomFactor   = 0.88
		baseFontSize = 35.0
		baseGap      = 60 * scale
		fontSize     = baseFontSize * scale
	)

	dc := gg.NewContext(int(canvasWidth), int(canvasHeight))
	fontBytes, err := ioutil.ReadFile("fonts/SFPRODISPLAYREGULAR.ttf")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	// parse font
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	// fill background with white
	dc.SetColor(color.White)
	dc.DrawRectangle(0, 0, canvasWidth, canvasHeight)
	dc.Fill()

	r, _ := regexp.Compile(`(?m)^.*\.(jpg|JPG|png|PNG|jpeg|JPEG)$`)

	if !r.MatchString(imagePayload.Filename) {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid file type")
	}

	fileExtension := r.FindStringSubmatch(imagePayload.Filename)[1] // example jpg, png
	fileName := fmt.Sprintf("%s.%s", time.Now().Format("20060102150405"), fileExtension)
	folderPath := "uploads"
	fileFullPath := fmt.Sprintf("%s/%s", folderPath, fileName)

	// check if folder exists
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		os.MkdirAll(folderPath, 0755)
	}

	// save file to disk
	if err := c.SaveFile(imagePayload, fileFullPath); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// get file extension from the file name

	et, err := exiftool.NewExiftool()
	if err != nil {
		log.Fatal(err)
	}
	defer et.Close()
	fileInfos := et.ExtractMetadata(fileFullPath)

	imageDecoded, err := imaging.Open(fileFullPath)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if rotateAngle != 0.0 {
		imageDecoded = imaging.Rotate(imageDecoded, rotateAngle, color.Transparent)
	}

	metadataObject := MetadataType{}
	metadataObject.DateTimeOriginal = time.Now().Format("2006:01:02 15:04:05")
	metadataObject.Model = "Unknown"
	metadataObject.FocalLengthIn35mmFormat = "0 mm"
	metadataObject.Aperture = 0.0
	metadataObject.ShutterSpeed = "Unknown"
	metadataObject.ISO = 0.0

	for _, fileInfo := range fileInfos {
		if fileInfo.Err != nil {
			fmt.Printf("Error concerning %v: %v\n", fileInfo.File, fileInfo.Err)
			continue
		}
		log.Print("\n\n")
		log.Printf("FileName: %v", fileInfo.File)
		log.Printf("DateTimeOriginal: %v", fileInfo.Fields["DateTimeOriginal"])
		log.Printf("CameraModelName: %v", fileInfo.Fields["CameraModelName"])
		log.Printf("Model: %v", fileInfo.Fields["Model"])
		log.Printf("FocalLengthIn35mmFormat: %v", fileInfo.Fields["FocalLengthIn35mmFormat"])
		log.Printf("FocalLength: %v", fileInfo.Fields["FocalLength"])
		log.Printf("Aperture: %v", fileInfo.Fields["Aperture"])
		log.Printf("ShutterSpeed: %v", fileInfo.Fields["ShutterSpeed"])
		log.Printf("ISO: %v", fileInfo.Fields["ISO"])

		if fileInfo.Fields["DateTimeOriginal"] == nil || fileInfo.Fields["Aperture"] == nil || fileInfo.Fields["ShutterSpeed"] == nil || fileInfo.Fields["ISO"] == nil {
			continue
		}

		metadataObject = MetadataType{
			DateTimeOriginal: fileInfo.Fields["DateTimeOriginal"].(string),
			Aperture:         fileInfo.Fields["Aperture"].(float64),
			ShutterSpeed:     fileInfo.Fields["ShutterSpeed"].(string),
			ISO:              fileInfo.Fields["ISO"].(float64),
		}

		if fileInfo.Fields["CameraModelName"] != nil {
			metadataObject.Model = fileInfo.Fields["CameraModelName"].(string)
		} else {
			metadataObject.Model = fileInfo.Fields["Model"].(string)
		}

		if fileInfo.Fields["FocalLengthIn35mmFormat"] != nil {
			metadataObject.FocalLengthIn35mmFormat = fileInfo.Fields["FocalLengthIn35mmFormat"].(string)
		} else {
			metadataObject.FocalLengthIn35mmFormat = fileInfo.Fields["FocalLength"].(string)
		}
	}

	latitude, longitude, err := services.GetCoordinate(fileFullPath)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	metadataObject.GPSLatitude = latitude
	metadataObject.GPSLongitude = longitude

	wrh := float64(imageDecoded.Bounds().Dx()) / float64(imageDecoded.Bounds().Dy())
	newWidth := canvasWidth * zoomFactor
	newHeight := newWidth / wrh
	if newHeight > canvasHeight {
		newHeight = canvasHeight
		newWidth = newHeight * wrh
	}

	xOffset := (canvasWidth - newWidth) / 2
	yOffset := (canvasHeight - newHeight) / 2

	imageDecoded = imaging.Resize(imageDecoded, int(newWidth), int(newHeight), imaging.Lanczos)
	dc.DrawImage(imageDecoded, int(xOffset), int(yOffset))

	// Draw image
	dc.SetColor(color.Black)
	dc.SetFontFace(truetype.NewFace(f, &truetype.Options{Size: fontSize}))
	dc.DrawString(fmt.Sprintf("Shot on %s @%s f/%.1f", metadataObject.Model, metadataObject.FocalLengthIn35mmFormat, metadataObject.Aperture), xOffset+6*scale, yOffset+newHeight+baseGap)
	dc.SetFontFace(truetype.NewFace(f, &truetype.Options{Size: fontSize * 0.8}))
	t, err := time.Parse("2006:01:02 15:04:05", metadataObject.DateTimeOriginal)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	location, locationErr := services.GetLocation(metadataObject.GPSLatitude, metadataObject.GPSLongitude)
	if locationErr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": locationErr.Error(),
		})
	}

	// dc.DrawString(fmt.Sprintf("%s%s", t.Format("Jan 2, 2006 15:04"), locationFormat), xOffset+6*scale, yOffset+newHeight+baseGap+45*scale)
	dc.DrawString(fmt.Sprintf("%s%s", t.Format("Jan 2, 2006"), location), xOffset+6*scale, yOffset+newHeight+baseGap+45*scale)
	imageResult := dc.Image()
	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, imageResult); err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	c.Response().Header.Set("Content-Type", "image/png")
	os.RemoveAll(fileFullPath)
	return c.SendStream(buffer)
}
