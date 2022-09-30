package main

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"
	"path"
	"sync"

	"github.com/nfnt/resize"
)

const maxThreads = 32

func main() {
	entries, err := os.ReadDir(os.Args[1])
	if err != nil {
		panic(err)
	}

	compressedFolder := path.Join(os.Args[1], `compressed`)

	if err := os.Mkdir(compressedFolder, os.ModeDir); err != nil {
		panic(err)
	}

	threads := make(chan struct{})

	for i := 0; i < maxThreads; i++ {
		go func() {
			threads <- struct{}{}
		}()
	}

	i := 0
	var wg sync.WaitGroup
	for _, entry := range entries {
		e := entry
		<-threads
		wg.Add(1)

		go func() {
			file, err := os.Open(path.Join(os.Args[1], e.Name()))
			if err != nil {
				panic(err)
			}

			img, _, err := image.Decode(file)
			if err != nil {
				panic(err)
			}

			height := img.Bounds().Max.Y - img.Bounds().Min.Y
			width := img.Bounds().Max.X - img.Bounds().Min.X

			smaller := height
			if width < height {
				smaller = width
			}

			if smaller > 1000 {
				factor := 1000 / float64(smaller)

				img = resize.Resize(uint(factor*float64(width)), uint(factor*float64(height)), img, resize.Bicubic)
			}

			outFile, err := os.Create(path.Join(compressedFolder, e.Name()))
			if err != nil {
				panic(err)
			}

			if err := jpeg.Encode(outFile, img, &jpeg.Options{Quality: 85}); err != nil {
				panic(err)
			}

			i++
			fmt.Printf("(%d/%d) %s\n", i, len(entries), e.Name())
			go func() {
				threads <- struct{}{}
			}()
			wg.Done()
		}()
	}

	wg.Wait()
}
