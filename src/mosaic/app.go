package main

import (
	"fmt"
	"html/template"
	"image"
	"image/color"
	//_ "image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	//"math"
	"log"
	"math"
	"mosaic/colordiff"
	"mosaic/imageutil"
	"mosaic/ini"
	"net/http"
	"os"
	"path/filepath"
	"reflect"

	/*
		"flag"
		"github.com/gorilla/mux"
		"log"

		"acquire/conn"
		"acquire/game"
		"acquire/ini"

		"database/sql"
		_ "github.com/go-sql-driver/mysql"*/)

/*var (
	addr = flag.Bool("addr", false, "find open address and print to final-port.txt")
)

var router = mux.NewRouter()
var config ini.Dict
var err error

func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ProductsHandler::Dashboard")
}*/

var config ini.Dict
var err error

func init() {
	// without this register .. At(), Bounds() functions will
	// caused memory pointer error!!
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	//image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
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
	Images        []PoolImage
	AverageAspect float64
}

func (p Pool) ComputeMse() (float64, error) {

	return 5.0, nil
}

func MakePool(path string) *Pool {
	dir_to_scan := "C:/tmp/uploadedfile/pool"
	files, _ := ioutil.ReadDir(dir_to_scan)

	MaxImage := 1

	//pool := make([]PoolImage, len(files))
	pool := make([]PoolImage, MaxImage)
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
			pool[key] = PoolImage{Name: imgFile.Name(), Width: bounds.Max.Y, Height: bounds.Max.X, Aspect: aspect, Image: imageData}
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

// TODO: make a array of pool image with aspect ratio
//  	FIX config.ini make it my code
//
//
type PoolImage struct {
	Name   string
	Width  int
	Height int
	Aspect float64
	Image  image.Image
}

func (img *PoolImage) Resize(w, h int) {
	img.Image = imageutil.Resize(img.Image, w, h, imageutil.Lanczos)
	img.Width = w
	img.Height = h
}

func main() {
	config, err = ini.Load("ini/config.ini")
	if err != nil {
		log.Fatal("Failed to load config")
	}

	// search model
	//model := search.NewModel()

	// POOL Images
	pool := MakePool("C:/tmp/uploadedfile/pool")

	// TARGET Image
	if reader, err := os.Open("C:/tmp/uploadedfile/HTML5_Logo_512.png"); err == nil {
		targetImg, _, err := image.Decode(reader)
		if err != nil {
			fmt.Println("err", err)
			return
		}
		fmt.Println(reflect.TypeOf(targetImg))
		//fmt.Printf("%d %d\n", targetImg.Width, targetImg.Height)
		bounds := targetImg.Bounds()
		targetWidth := float64(bounds.Max.X)
		targetHeight := float64(bounds.Max.Y)
		target_aspect := targetWidth / targetHeight
		fmt.Println("target_aspect:", target_aspect)

		// size
		//thumb_aspect = sum(game.aspect for game in games) / len(games)
		//patchHeight = int(float(patch_w) * thumb_aspect)
		// TODO: patch size should be divisible by target width; aka round number of patches for row/column
		//       should patch size be nearest square root of image size?
		// TODO: resize target to match patch sizes also maintain aspect ratio
		pw, found := config.GetInt("options", "patchwidth")
		if !found {
			log.Fatal("Couldn't get patchWidth")
		}
		patchWidth := float64(pw)
		// patchHeight based on average of pool
		patchHeight := float64(patchWidth) * pool.AverageAspect
		fmt.Println("Patch:", patchWidth, patchHeight)

		target := targetImg.At(200, 200)
		fmt.Println("target:", target)
		palette := make([]color.Color, 10)
		for i := 0; i < 10; i++ {
			palette[i] = targetImg.At(i+100, i+100)
		}
		fmt.Println(palette)
		x := colordiff.Closest(target, palette)
		fmt.Println("Closest:", x)

		/*for _, poolImage := range pool.Images {
			poolImage.Resize(int(patchWidth), int(patchHeight))
			colorWord := search.WordFromImage(poolImage.Image)
			model.TrainWord(colorWord)
		}*/

		/*cols, rows := 1.0, 1.0
		poolLen := float64(len(pool.Images))
		for cols*rows < poolLen {
			col_asp := (cols + 1) * patchWidth / (math.Ceil(poolLen/cols+1) * patchHeight)
			row_asp := cols * patchWidth / (math.Ceil(poolLen/cols) * patchHeight)
			if math.Abs(col_asp-target_aspect) < math.Abs(row_asp-target_aspect) {
				cols++
			} else {
				rows++
			}
		}
		newTargetW := int(cols * patchWidth)
		newTargetH := int(rows * patchHeight)

		// adjust target
		adjustedTarget := imageutil.Resize(targetImg, newTargetW, newTargetH, imageutil.Lanczos)
		targetData := make([]color.Color, newTargetW*newTargetH)
		for y := 0.0; y < rows; y++ {
			yOfset := y * patchHeight
			for x := 0.0; x < cols; x++ {
				xOfset := x * patchHeight
				for yy := 0.0; yy < patchHeight; yy++ {
					for xx := 0.0; xx < patchWidth; xx++ {
						rgbaPix := adjustedTarget.At(int(xOfset+xx), int(yOfset+yy))
						targetData = append(targetData, rgbaPix)
					}
				}

			}
		}*/
		// pool MSE
		//pool.ComputeMse()

		// test stuff

		// test color

		/*
			// test output
			out, err := os.Create("C:/tmp/uploadedfile/HTML5_Logo_512___OUT.png")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err = png.Encode(out, adjustedTarget)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		*/
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
