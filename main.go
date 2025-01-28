package main

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
)

type MetadataType struct {
	DateTimeOriginal        string
	Model                   string
	FocalLengthIn35mmFormat string
	Aperture                float64
	ShutterSpeed            string
	ISO                     float64
}

func main() {
	app := fiber.New()

	app.Get("/image", func(c *fiber.Ctx) error {
		imagePayload, err := c.FormFile("image")

		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		rotateAngle := 0.0
		rotateAngle, err = strconv.ParseFloat(c.FormValue("rotateAngle"), 64)

		scale := 1.0
		scale, err = strconv.ParseFloat(c.FormValue("scale"), 64)

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
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		// parse font
		f, err := truetype.Parse(fontBytes)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		// fill background with white
		dc.SetColor(color.White)
		dc.DrawRectangle(0, 0, canvasWidth, canvasHeight)
		dc.Fill()

		r, _ := regexp.Compile(`(?m)^.*\.(jpg|JPG|png|PNG)$`)

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
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		// get file extension from the file name

		et, err := exiftool.NewExiftool()
		if err != nil {
			log.Fatal(err)
		}
		defer et.Close()
		fileInfos := et.ExtractMetadata(fileFullPath)

		imageDecoded, err := imaging.Open(fileFullPath)
		imageDecoded = imaging.Rotate(imageDecoded, rotateAngle, color.Transparent)

		metadataObject := MetadataType{}

		for _, fileInfo := range fileInfos {
			if fileInfo.Err != nil {
				fmt.Printf("Error concerning %v: %v\n", fileInfo.File, fileInfo.Err)
				continue
			}

			metadataObject = MetadataType{
				DateTimeOriginal:        fileInfo.Fields["DateTimeOriginal"].(string),
				Model:                   fileInfo.Fields["Model"].(string),
				FocalLengthIn35mmFormat: fileInfo.Fields["FocalLengthIn35mmFormat"].(string),
				Aperture:                fileInfo.Fields["Aperture"].(float64),
				ShutterSpeed:            fileInfo.Fields["ShutterSpeed"].(string),
				ISO:                     fileInfo.Fields["ISO"].(float64),
			}
		}
		fmt.Println(metadataObject)

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
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}
		dc.DrawString(fmt.Sprintf("%s", t.Format("Jan 2, 2006 15:04")), xOffset+6*scale, yOffset+newHeight+baseGap+45*scale)
		imageResult := dc.Image()
		buffer := new(bytes.Buffer)
		if err := png.Encode(buffer, imageResult); err != nil {
			c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		c.Response().Header.Set("Content-Type", "image/png")
		os.RemoveAll(fileFullPath)
		return c.SendStream(buffer)
	})

	log.Fatal(app.Listen(":3000"))
	log.Println("Server is running on port 3000")
}
