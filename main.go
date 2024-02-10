package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nfnt/resize"
)

var (
	targetWidth   uint
	targetHeight  uint
	jpgExtensions = ".jpg"
	pngExtensions = ".png"
)

func main() {
	// Define command-line flags
	inputDir := flag.String("input", "input_images", "Input directory containing images")
	outputDir := flag.String("output", "output_images", "Output directory for resized images")
	flag.UintVar(&targetWidth, "width", 0, "Target width for resizing")
	flag.UintVar(&targetHeight, "height", 0, "Target height for resizing")
	flag.Parse()

	// Check if required flags are provided
	if *inputDir == "" || *outputDir == "" || targetWidth == 0 || targetHeight == 0 {
		fmt.Println("Error: Please provide valid values for input directory, output directory, width, and height.")
		flag.PrintDefaults()
		return
	}

	// Read the input directory
	files, err := ioutil.ReadDir(*inputDir)
	if err != nil {
		fmt.Println("Error reading input directory:", err)
		return
	}

	if err := os.MkdirAll(*outputDir, os.ModePerm); err != nil {
		fmt.Println("Error creating output directory:", err)
		return
	}

	var wg sync.WaitGroup
	done := make(chan struct{})

	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(strings.ToLower(file.Name()), jpgExtensions) || strings.HasSuffix(strings.ToLower(file.Name()), pngExtensions)) {
			wg.Add(1)

			go func(file os.FileInfo) {
				defer wg.Done()

				inputPath := filepath.Join(*inputDir, file.Name())
				outputPath := filepath.Join(*outputDir, file.Name())

				err := resizeImage(inputPath, outputPath, int(targetWidth), int(targetHeight))
				if err != nil {
					fmt.Printf("Error resizing image %s: %v\n", file.Name(), err)
				} else {
					fmt.Printf("Successfully resized: %s\n", file.Name())
				}

			}(file)
		}
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	<-done
}

func resizeImage(inputPath, outputPath string, width, height int) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	// Use int values for width and height
	resizedImg := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Check if the file has a PNG extension
	if strings.HasSuffix(strings.ToLower(outputPath), pngExtensions) {
		err = png.Encode(outFile, resizedImg)
	} else {
		err = jpeg.Encode(outFile, resizedImg, nil)
	}

	if err != nil {
		return err
	}

	return nil
}

