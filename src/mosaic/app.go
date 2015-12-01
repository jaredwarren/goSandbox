package main

import (
	"fmt"
	"html/template"
	"image"
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
	//"reflect"
	"sync"
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

func MakePool(path string) *Pool {
	dir_to_scan := "C:/tmp/uploadedfile/pool"
	files, _ := ioutil.ReadDir(dir_to_scan)

	MaxImage := 100
	/*
		MaxImage := len(files)
	*/

	pool := make([]*PoolImage, MaxImage)
	total_aspect := float64(0)
	counter := float64(0)
	for key, imgFile := range files {
		poolImage := loadFile(filepath.Join(dir_to_scan, imgFile.Name()))
		total_aspect += poolImage.Aspect
		pool[key] = poolImage

		counter++
		if counter >= float64(MaxImage) {
			break
		}
	}
	return &Pool{Images: pool, AverageAspect: total_aspect / counter}
}

func loadFile(filePath string) *PoolImage {
	if reader, err := os.Open(filePath); err == nil {
		defer reader.Close()
		imageData, imageType, err := image.Decode(reader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", filePath, err)
			return nil
		}
		if imageType == "png" {

		}
		bounds := imageData.Bounds()
		aspect := float64(bounds.Max.Y) / float64(bounds.Max.X)
		return &PoolImage{Name: filePath, Width: bounds.Max.Y, Height: bounds.Max.X, Aspect: aspect, Image: imageData}

	} else {
		fmt.Println("Impossible to open the file:", err)
	}
	if err != nil {
		panic(err)
	}
	return nil
}

func Round(f float64) int {
	return int(math.Floor(f + .5))
}

type PoolImage struct {
	Name   string
	Width  int
	Height int
	Aspect float64
	Image  image.Image
	Colors []colordiff.LAB
	Deltas map[string]float64
}

func (img *PoolImage) Resize(w, h int) {
	if w != img.Width || h != img.Height {
		img.Image = imageutil.Resize(img.Image, w, h, imageutil.Lanczos)
		img.Width = w
		img.Height = h
	}
}
func (img *PoolImage) CacheColors() {
	img.Colors = []colordiff.LAB{}
	for y := 0; y < img.Width; y += sampleSize {
		for x := 0; x < img.Height; x += sampleSize {
			img.Colors = append(img.Colors, colordiff.RgbToLab(img.Image.At(x, y)))
		}
	}
}

//TODO: I'm going to try to calculate the average color in LAB for each patch and pool image, see if that's faster and still accurate

// TODO: cache deltas
//
//
// Calculate average error for each patch
func (img *PoolImage) CalculateError(targetPatch []colordiff.LAB) float64 {
	if len(targetPatch) != len(img.Colors) {
		fmt.Println("ERROR:", len(targetPatch), "!=", len(img.Colors))
	}

	totalDiff := 0.0
	for i := 0; i < len(targetPatch); i++ {
		totalDiff += colordiff.Diff2(targetPatch[i], img.Colors[i])
	}

	return totalDiff / float64(len(targetPatch))
}

var sampleSize int = 1

func main() {
	config, err = ini.Load("ini/config.ini")
	if err != nil {
		log.Fatal("Failed to load config")
	}

	ss, found := config.GetInt("options", "samplesize")
	if !found {
		log.Fatal("Couldn't get sampleSize")
	}
	sampleSize = ss
	fmt.Println("Sample Size:", sampleSize)

	// POOL Images
	pool := MakePool("C:/tmp/uploadedfile/pool")
	fmt.Println("Pool Size:", len(pool.Images))

	// TARGET Image
	if reader, err := os.Open("C:/tmp/uploadedfile/HTML5_Logo_512.png"); err == nil {
		targetImg, _, err := image.Decode(reader)
		if err != nil {
			fmt.Println("err", err)
			return
		}
		// patch size
		patchWidth, found := config.GetInt("options", "patchwidth")
		if !found {
			log.Fatal("Couldn't get patchWidth")
		}
		// patchHeight based on average of pool
		patchHeight := Round(float64(patchWidth) * pool.AverageAspect)
		fmt.Println("Patch:", patchWidth, patchHeight)
		// resize pool
		for _, poolImage := range pool.Images {
			poolImage.Resize(int(patchWidth), int(patchHeight))
			poolImage.CacheColors()
		}

		// adjust target
		bounds := targetImg.Bounds()
		targetWidth := bounds.Max.X
		targetHeight := bounds.Max.Y
		target_aspect := float64(targetWidth) / float64(targetHeight)
		cols, rows := 1, 1
		poolLen := len(pool.Images)
		for cols*rows < poolLen {
			colAspect := (float64(cols + 1)) * float64(patchWidth) / (math.Ceil(float64(poolLen)/float64(cols+1)) * float64(patchHeight))
			rowAspect := float64(cols*patchWidth) / (math.Ceil(float64(poolLen)/float64(cols)) * float64(patchHeight))
			if math.Abs(colAspect-target_aspect) < math.Abs(rowAspect-target_aspect) {
				cols++
			} else {
				rows++
			}
		}
		fmt.Println("cols:", cols, "rows:", rows)
		newTargetW := int(cols * patchWidth)
		newTargetH := int(rows * patchHeight)
		fmt.Println("New Target Size:", newTargetW, newTargetH)
		adjustedTarget := imageutil.Resize(targetImg, newTargetW, newTargetH, imageutil.Lanczos)

		//TODO: see if I can't use GO Routines

		//compare with pool
		t0 := time.Now()
		percentCounter := 0.0
		outImage := image.NewRGBA(image.Rect(0, 0, newTargetW, newTargetH))
		/*// resize pool
		var wg sync.WaitGroup
		for _, poolImage := range pool.Images {
			wg.Add(1)
			go func(pi *PoolImage) {
				defer wg.Done()
				pi.Resize(int(patchWidth), int(patchHeight))
				pi.CacheColors()
			}(poolImage)
		}
		wg.Wait()*/
		for row := 0; row < rows; row++ {
			for col := 0; col < cols; col++ {
				patch := make([]colordiff.LAB, int(math.Ceil(float64(patchWidth)/float64(sampleSize))*math.Ceil(float64(patchHeight)/float64(sampleSize))))
				xOfset := col * patchWidth
				yOfset := row * patchHeight
				iCounter := 0
				for y := 0; y < patchHeight; y += sampleSize {
					for x := 0; x < patchWidth; x += sampleSize {
						patch[iCounter] = colordiff.RgbToLab(adjustedTarget.At(xOfset+x, yOfset+y))
						iCounter++
					}
				}

				// find best match
				var bestPoolImage *PoolImage
				lowestDelta := math.MaxFloat64
				for _, poolImage := range pool.Images {
					delta := poolImage.CalculateError(patch)
					if delta < lowestDelta {
						lowestDelta = delta
						bestPoolImage = poolImage
					}
				}
				p := image.Pt(-col*int(patchWidth), -row*int(patchHeight))
				draw.Draw(outImage, outImage.Bounds(), bestPoolImage.Image, p, draw.Src)
				// percent
				percentCounter++
				fmt.Print("\r", int((percentCounter/float64(rows*cols))*100), "%")

			}
		}
		fmt.Println("")
		fmt.Printf("Time: %v\n", time.Now().Sub(t0))
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
