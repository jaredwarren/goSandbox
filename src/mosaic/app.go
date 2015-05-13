package main

import (
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"math"
	"mosaic/colordiff"
	"mosaic/imageutil"
	"mosaic/ini"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

var config ini.Dict
var err error

func init() {
	// without this register .. At(), Bounds() functions will
	// caused memory pointer error!!
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("~uploadHandler~")

	// the FormFile function takes in the POST input id file
	file, header, err := r.FormFile("file")
	if err == nil {
		fmt.Fprintln(w, err)
		return
	}

	defer file.Close()

	// verify file type
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	filetype := http.DetectContentType(buff)
	switch filetype {
	case "image/png":
		break
	case "image/jpeg", "image/jpg":
	case "image/gif":
	default:
		fmt.Println(err)
		fmt.Fprintf(w, "Invalid file type uploaded")
		return
	}

	// TODO: clean/verify filename
	imgfile, err := os.Create("C:/tmp/uploadedfile/" + header.Filename)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
		return
	}

	defer imgfile.Close()

	// write the content from POST to the file
	_, err = io.Copy(imgfile, file)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	// Process image
	//imgfileX, err := os.Open("C:/tmp/uploadedfile/" + header.Filename)
	img, _, err := image.Decode(imgfile)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	//defer imgfileX.Close()

	bounds := img.Bounds()
	fmt.Println(bounds)

	fmt.Fprintf(w, "File uploaded successfully : ")
	fmt.Fprintf(w, header.Filename)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("indexHandler")

	// for now parse every request so I don't have to recompile, maybe
	var tmpl = template.Must(template.ParseFiles("static/templates/index.html", "static/templates/base.html"))
	/*pagedata := &common.Page{Tags: &common.Tags{Id: 1, Name: "golang"},
	Content: &common.Content{Id: 9, Title: "Hello", Content: "World!"},
	Comment: &common.Comment{Id: 2, Note: "Good Day!"}}*/

	tmpl.ExecuteTemplate(w, "base", "")
}

type Pool struct {
	Images        []*PoolImage
	AverageAspect float64
}

func (p Pool) ComputeMse() (float64, error) {

	return 5.0, nil
}

func MakePool(path string) *Pool {
	dir_to_scan := "C:/tmp/uploadedfile/pool"
	files, _ := ioutil.ReadDir(dir_to_scan)

	MaxImage := 200
	/*
		MaxImage := len(files)

	*/

	//pool := make([]PoolImage, len(files))
	pool := make([]*PoolImage, MaxImage)
	total_aspect := float64(0)
	counter := float64(0)
	for key, imgFile := range files {
		if reader, err := os.Open(filepath.Join(dir_to_scan, imgFile.Name())); err == nil {
			defer reader.Close()
			imageData, imageType, err := image.Decode(reader)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", imgFile.Name(), err)
				continue
			}
			if imageType == "png" {

			}
			// TODO: resize to fit patch size
			bounds := imageData.Bounds()
			aspect := float64(bounds.Max.Y) / float64(bounds.Max.X)
			total_aspect += aspect
			counter++
			pool[key] = &PoolImage{Name: imgFile.Name(), Width: bounds.Max.Y, Height: bounds.Max.X, Aspect: aspect, Image: imageData}
			if counter >= float64(MaxImage) {
				break
			}
		} else {
			fmt.Println("Impossible to open the file:", err)
		}
	}
	if err != nil {
		panic(err)
	}
	return &Pool{Images: pool, AverageAspect: total_aspect / counter}
}

func Round(f float64) float64 {
	return math.Floor(f + .5)
}

type PoolImage struct {
	Name   string
	Width  int
	Height int
	Aspect float64
	Image  image.Image
	Deltas map[string]float64
}

func (img *PoolImage) Resize(w, h int) {
	if w != img.Width || h != img.Height {
		img.Image = imageutil.Resize(img.Image, w, h, imageutil.Lanczos)
		img.Width = w
		img.Height = h
	}
}

//TODO: I'm going to try to calculate the average color in LAB for each patch and pool image, see if that's faster and still accurate

// TODO: cache deltas
//
//
// Calculate average error for each patch
func (img *PoolImage) CalculateError(image image.Image, col, row int, patchWidth, patchHeight int) float64 {
	/*if len(targetPatch) != img.Width*img.Height {
		fmt.Println("ERROR:", len(targetPatch), "!=", img.Width*img.Height)
	}*/

	totalDiff := 0.0
	//yIndex := 0
	//xIndex := 0
	//width := float64(img.Width)
	xOfset := col * patchWidth
	yOfset := row * patchHeight
	//c1 := color.NRGBA{0, 0, 0, 0}
	for y := 0; y < patchHeight; y++ {
		for x := 0; x < patchWidth; x++ {
			//yIndex = yOfset + y
			//xIndex = xOfset + x
			totalDiff += colordiff.Diff(image.At(xOfset+x, yOfset+y), img.Image.At(x, y))
			//totalDiff += colordiff.Diff(c1, c1)
		}
	}
	/*for pixelIndex, targetColor := range targetPatch {
		yIndex = int(math.Floor(float64(pixelIndex) / width))
		xIndex = pixelIndex - (yIndex * img.Width)
		totalDiff += colordiff.Diff(targetColor, img.Image.At(xIndex, yIndex))
	}*/
	//fmt.Println(totalDiff / float64(patchHeight*patchWidth))
	return totalDiff / float64(patchHeight*patchWidth)
}

func main() {
	config, err = ini.Load("ini/config.ini")
	if err != nil {
		log.Fatal("Failed to load config")
	}

	// POOL Images
	t0 := time.Now()
	pool := MakePool("C:/tmp/uploadedfile/pool")
	fmt.Println("Pool:", len(pool.Images))
	fmt.Printf("Make Pool: %v\n", time.Now().Sub(t0))

	// TARGET Image
	if reader, err := os.Open("C:/tmp/uploadedfile/HTML5_Logo_512.png"); err == nil {
		targetImg, _, err := image.Decode(reader)
		if err != nil {
			fmt.Println("err", err)
			return
		}
		// patch size
		pw, found := config.GetInt("options", "patchwidth")
		if !found {
			log.Fatal("Couldn't get patchWidth")
		}
		patchWidth := float64(pw)
		// patchHeight based on average of pool
		patchHeight := float64(patchWidth) * pool.AverageAspect
		fmt.Println("Patch:", patchWidth, patchHeight)
		// resize pool
		for _, poolImage := range pool.Images {
			poolImage.Resize(int(patchWidth), int(patchHeight))
		}

		// adjust target
		bounds := targetImg.Bounds()
		targetWidth := float64(bounds.Max.X)
		targetHeight := float64(bounds.Max.Y)
		target_aspect := targetWidth / targetHeight
		cols, rows := 1.0, 1.0
		poolLen := float64(len(pool.Images))
		for cols*rows < poolLen {
			col_asp := (cols + 1) * patchWidth / (math.Ceil(poolLen/(cols+1)) * patchHeight)
			row_asp := cols * patchWidth / (math.Ceil(poolLen/cols) * patchHeight)
			if math.Abs(col_asp-target_aspect) < math.Abs(row_asp-target_aspect) {
				cols++
			} else {
				rows++
			}
		}
		//cols := 50.0
		//rows := 50.0
		//cols /= 2
		//rows /= 2
		fmt.Println("cols:", cols, "rows:", rows)
		newTargetW := int(cols * patchWidth)
		newTargetH := int(rows * patchHeight)
		adjustedTarget := imageutil.Resize(targetImg, newTargetW, newTargetH, imageutil.Lanczos)
		fmt.Println(reflect.TypeOf(adjustedTarget), newTargetW, newTargetH)

		t0 := time.Now()
		counter := 0.0

		//targetPatches := [][]color.Color{}
		//imagePools := make([]*PoolImage, int(cols*rows))
		outImage := image.NewRGBA(image.Rect(0, 0, newTargetW, newTargetH))
		for row := 0; row < int(rows); row++ {
			for col := 0; col < int(cols); col++ {
				// find best match
				var bestPoolImage *PoolImage
				lowestDelta := math.MaxFloat64
				for _, poolImage := range pool.Images {
					delta := poolImage.CalculateError(adjustedTarget, col, row, int(patchWidth), int(patchHeight))
					if delta < lowestDelta {
						lowestDelta = delta
						bestPoolImage = poolImage
					}
				}
				p := image.Pt(-col*int(patchWidth), -row*int(patchHeight))
				draw.Draw(outImage, outImage.Bounds(), bestPoolImage.Image, p, draw.Src)
				// percent
				counter++
				fmt.Println(int((counter/(rows*cols))*100), "%")

			}
		}
		fmt.Printf("Time: %v\n", time.Now().Sub(t0))
		// calculate diff
		// TODO: move this inside previous for loop?

		/*for _, poolImage := range pool.Images {
			poolImage.Resize(int(patchWidth), int(patchHeight))
			poolImage.CalculateError(targetPatches)
		}*/

		// Position pool

		/*for i, _ := range targetPatches {
			var bestPoolImage *PoolImage
			lowestError := math.MaxFloat64
			for _, poolImage := range pool.Images {
				if poolImage.Deltas[i] < lowestError {
					bestPoolImage = poolImage
					lowestError = poolImage.Deltas[i]
				}
			}
			yIndex = int(math.Floor(float64(i) / cols))
			xIndex = i - (yIndex * int(cols))
			p := image.Pt(-xIndex*int(patchWidth), -yIndex*int(patchHeight))
			draw.Draw(outImage, outImage.Bounds(), bestPoolImage.Image, p, draw.Src)
			imagePools[i] = bestPoolImage
		}*/

		/*
		 */
		// test output
		out, err := os.Create("C:/tmp/uploadedfile/HTML5_Logo_512___OUT.png")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = png.Encode(out, outImage)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return

		//fmt.Println(cw)

		// for each patch
		// 		resize to match patch
		//      foreach row
		// 			foreach col
		// 				append to list/word
		//
		//
		// foreach patch in target
		//      foreach row
		// 			foreach col
		// 				append to list/word

		// ??? Can I change this so i resize the target, but not the patches/pool images?

		return

		// TODO: resize target to xPaches

		//fmt.Println(reflect.TypeOf(bounds.Max.Y))
		//fmt.Println(patchWidth, patchHeight)
		//y_patches := float64(bounds.Max.Y) / float64(patchHeight)
		//x_patches := float64(bounds.Max.X) / float64(patchWidth)
		//fmt.Println(x_patches, y_patches)

		//x_patches := bounds.Max.X / patchWidth

		/*
		 */

		/*var histogram [16][4]int
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, a := img.At(x, y).RGBA()
				// A color's RGBA method returns values in the range [0, 65535].
				// Shifting by 12 reduces this to the range [0, 15].
				histogram[r>>12][0]++
				histogram[g>>12][1]++
				histogram[b>>12][2]++
				histogram[a>>12][3]++
			}
		}
		// Print the results.
		fmt.Printf("%-14s %6s %6s %6s %6s\n", "bin", "red", "green", "blue", "alpha")
		for i, x := range histogram {
			fmt.Printf("0x%04x-0x%04x: %6d %6d %6d %6d\n", i<<12, (i+1)<<12-1, x[0], x[1], x[2], x[3])
		}*/
		//fmt.Println(bounds)
	} else {
		fmt.Println(err)
		return
	}

	return

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/receive", uploadHandler)

	fmt.Println("Running....")
	http.ListenAndServe(":8080", nil)

	/*config, err = ini.Load("ini/config.ini")
	if err != nil {
		log.Fatal("Failed to load config")
	}

	// Routs
	r := router
	//r.HandleFunc("/static/{path:.*}", common.StaticHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/game/", game.MakeMuxer("/game/"))

	// websocket
	http.Handle("/ws/", conn.MakeMuxer("/ws/"))

	// wait for clients
	http.Handle("/", r)
	fmt.Println("Running...\n")
	http.ListenAndServe(":8080", nil)*/
}

func getPatchData(image image.Image, col, row int, patchWidth, patchHeight int) []color.Color {
	//patchData := make([]color.Color, patchWidth*patchHeight)
	patchData := []color.Color{}
	xOfset := col * patchWidth
	yOfset := row * patchHeight
	for y := 0; y < patchHeight; y++ {
		for x := 0; x < patchWidth; x++ {
			rgbaPix := image.At(int(xOfset+x), int(yOfset+y))
			patchData = append(patchData, rgbaPix)
		}
	}
	return patchData
}
