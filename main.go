package main

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
)

const (
	IMAGES_DIR           = "images"
	OUTPUT_DIR           = "output"
	MAX_PIXEL_DIFF_RATIO = 40
	DESIRED_DIMENSION    = 32
)

func main() {
	start := time.Now()
	files, err := ioutil.ReadDir(IMAGES_DIR)
	if err != nil {
		log.Fatal(err)
	}
	os.Mkdir(OUTPUT_DIR, os.ModePerm)

	count := work(files, IMAGES_DIR, OUTPUT_DIR)
	for _, file := range files {
		if file.IsDir() {
			fileName := file.Name()
			inputPath := path.Join(IMAGES_DIR, fileName)
			outputPath := path.Join(OUTPUT_DIR, fileName)
			fileFiles, err := ioutil.ReadDir(inputPath)
			if err == nil {
				os.Mkdir(outputPath, os.ModePerm)
				workCount := work(fileFiles, inputPath, outputPath)
				fmt.Printf("Completed '%s' with %v images\n", inputPath, workCount)
				count += workCount
			}
		}
	}
	fmt.Printf("Processed %v images in %s", count, time.Since(start))
}

func work(files []os.FileInfo, inputPath, outputPath string) int {
	count := 0

	for _, fileInfo := range files {
		fileName := fileInfo.Name()
		if fileInfo.IsDir() {
			fmt.Printf("Skipped directory '%s'", fileName)
			continue
		}

		// open file
		file, err := os.Open(path.Join(inputPath, fileName))
		if err != nil {
			log.Fatal(err)
		}

		// decode jpeg into image.Image
		img, err := jpeg.Decode(file)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()

		processedImage, err := processImage(fileName, img)
		if err != nil {
			continue
		}

		out, err := os.Create(path.Join(outputPath, fileName))
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()

		// write new image to file
		jpeg.Encode(out, processedImage, nil)
		count++
	}

	return count
}

func processImage(imageName string, img image.Image) (image.Image, error) {
	//start := time.Now()
	// get size
	size := CreatePoint(img.Bounds().Max)
	//fmt.Printf("%s: [%v,%v]", imageName, size.X, size.Y)

	pixelDiff := int((float32(size.GetMax())/float32(size.GetMin()) - 1) * 100)
	if pixelDiff > MAX_PIXEL_DIFF_RATIO {
		//	fmt.Printf(" Rejected, %v%% dimensions difference [%s]\n", pixelDiff, time.Since(start))
		return nil, errors.New("Invalid dimensions")
	}

	// crop
	if !size.Equal() {
		var anchor image.Point
		toCrop := (size.GetMax() - size.GetMin()) / 2
		if size.X < size.Y {
			anchor = image.Point{0, toCrop}
		} else {
			anchor = image.Point{toCrop, 0}
		}
		var err error
		img, err = cutter.Crop(img, cutter.Config{
			Width:  size.GetMin(),
			Height: size.GetMin(),
			Anchor: anchor,
			Mode:   cutter.TopLeft,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	// resize (0 to preserve)
	if size.X != DESIRED_DIMENSION || size.Y != DESIRED_DIMENSION {
		img = resize.Resize(DESIRED_DIMENSION, DESIRED_DIMENSION, img, resize.Lanczos3)
	}

	// grayscale
	grayImg := image.NewGray(img.Bounds())
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			grayImg.Set(x, y, img.At(x, y))
		}
	}

	//fmt.Printf(" Processed [%s]\n", time.Since(start))
	return grayImg, nil
}
