package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/karmdip-mi/go-fitz"
	"github.com/signintech/gopdf"
	"image"
	"image/draw"
	"image/jpeg"
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

func a3TOa4(fileName string, reduce int) []string{
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

	imgWidth := imc.Width - reduce
	halfWidth := imgWidth/2
	for i:= 0;i<=1;i++ {
		newImg  := image.NewRGBA(image.Rect(0,0, halfWidth,imc.Height))
		left := halfWidth * i + (reduce/2 * i * -1)
		fmt.Println("left: ",i,left)
		r := &image.Point{X: left  ,Y:0}
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

var in string
var out string
func init() {
	flag.StringVar(&in,"in","","输入文件名")
	flag.StringVar(&out,"out","","输出文件名")
}
func main() {
	flag.Parse()

	if in == "" || out == "" {
		log.Println("输入/输出文件名不能为空")
		os.Exit(1)
	}

	images := pdfToImage2(in)

	pdf := gopdf.GoPdf{}
	rec := gopdf.Rect{
		W: 1480,
		H: 2080,
	}
	pdf.Start(gopdf.Config{PageSize: rec})
	fmt.Println(rec)
	for idx, imageName := range images {
		left := 0
		switch idx {
		case 0:
			left=6
		case 1:
			left = 60
		case 3:
			left=140

		}

		fmt.Println(imageName,left)
		names := a3TOa4(imageName, left)
		for _,name := range names {
			pdf.AddPage()
			img,_ := ioutil.ReadFile(name)
			imc, _, _ := image.DecodeConfig(bytes.NewReader(img))
			fmt.Println(imc.Width)
			fmt.Println(imc.Height)
			pdf.Image(name, 0, 0, nil) //print image
		}
	}

	pdf.WritePdf(out)

}

