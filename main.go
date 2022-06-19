package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/karmdip-mi/go-fitz"
	"github.com/signintech/gopdf"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)



func pdfToImage2(file string) []string{
	doc, err := fitz.New(file)
	if err != nil {
		panic(err)
	}
	folder := strings.TrimSuffix(path.Base(file), filepath.Ext(path.Base(file)))

	imageNames := make([]string,0)
	// Extract pages as images
	for n := 0; n < doc.NumPage(); n++ {
		img, err := doc.Image(n)
		if err != nil {
			panic(err)
		}
		err = os.MkdirAll("img/"+folder, 0755)
		if err != nil {
			panic(err)
		}
		imageName := filepath.Join("img/"+folder+"/", fmt.Sprintf("image-%05d.jpg", n))
		imageNames = append(imageNames,imageName)

		f, err := os.Create(imageName )
		if err != nil {
			panic(err)
		}

		err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
		if err != nil {
			panic(err)
		}

		f.Close()
	}
	return imageNames
}

func a3TOa4(fileName string) []string{
	newImageFileNames := make([]string,0)
	file,err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	data,_ := ioutil.ReadAll(file)
	// Decoding gives you an Image.
	// If you have an io.Reader already, you can give that to Decode
	// without reading it into a []byte.
	img, _, _ := image.Decode(bytes.NewReader(data))
	imc, _, _ := image.DecodeConfig(bytes.NewReader(data))

	fmt.Println(imc.Width,imc.Height,img.Bounds())

	halfWidth := imc.Width/2
	for i:= 0;i<=1;i++ {
		newImg  := image.NewRGBA(image.Rect(0,0, halfWidth,imc.Height))
		r := &image.Point{X: halfWidth * i ,Y:0}
		if i == 1 {
			r.X -= 0
		}
		draw.Draw(newImg, newImg.Bounds(), img,*r, draw.Src)

		newFileName := fileName + fmt.Sprintf("-%v.jpg",i)
		newImageFileNames = append(newImageFileNames,newFileName)
		f, err := os.Create(newFileName)
		if err != nil {
			panic(err)
		}
		de := jpeg.Encode(f, newImg, nil)

		if de != nil {
			panic(de)
		}
	}
	return newImageFileNames
}

func main() {
	images := pdfToImage2("1.pdf")

	pdf := gopdf.GoPdf{}
	rec := gopdf.Rect{
		W: 1480,
		H: 2080,
	}
	pdf.Start(gopdf.Config{PageSize: rec})
	fmt.Println(rec)
	for _, imageName := range images {
		names := a3TOa4(imageName)
		for _,name := range names {
			pdf.AddPage()
			img,_ := ioutil.ReadFile(name)
			imc, _, _ := image.DecodeConfig(bytes.NewReader(img))
			fmt.Println(imc.Width)
			fmt.Println(imc.Height)
			pdf.Image(name, 70, 0, nil) //print image
		}
	}

	pdf.WritePdf("a4.pdf")

}


func  Trimming(sourceFileName string, destFileName string, x, y, w, h int) {
	src, err := LoadImage(sourceFileName)
	if err != nil {
		log.Println("load image fail..")
	}

	img, err := ImageCopy(src, x, y, w, h)
	if err != nil {
		log.Println("image copy fail...")
	}
	saveErr := SaveImage(destFileName, img)
	if saveErr != nil {
		log.Println("save image fail..")
	}
}

func  LoadImage(path string) (img image.Image, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	img, _, err = image.Decode(file)
	return
}

func  SaveImage(p string, src image.Image) error {
	f, err := os.OpenFile(p, os.O_SYNC|os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		return err
	}
	defer f.Close()
	ext := filepath.Ext(p)

	if strings.EqualFold(ext, ".jpg") || strings.EqualFold(ext, ".jpeg") {

		err = jpeg.Encode(f, src, &jpeg.Options{Quality: 80})

	} else if strings.EqualFold(ext, ".png") {
		err = png.Encode(f, src)
	} else if strings.EqualFold(ext, ".gif") {
		err = gif.Encode(f, src, &gif.Options{NumColors: 256})
	}
	return err
}

func ImageCopy(src image.Image, x, y, w, h int) (image.Image, error) {

	var subImg image.Image

	if rgbImg, ok := src.(*image.YCbCr); ok {
		subImg = rgbImg.SubImage(image.Rect(x, y, x+w, y+h)).(*image.YCbCr) //图片裁剪x0 y0 x1 y1
	} else if rgbImg, ok := src.(*image.RGBA); ok {
		subImg = rgbImg.SubImage(image.Rect(x, y, x+w, y+h)).(*image.RGBA) //图片裁剪x0 y0 x1 y1
	} else if rgbImg, ok := src.(*image.NRGBA); ok {
		subImg = rgbImg.SubImage(image.Rect(x, y, x+w, y+h)).(*image.NRGBA) //图片裁剪x0 y0 x1 y1
	} else {

		return subImg, errors.New("图片解码失败")
	}

	return subImg, nil
}
