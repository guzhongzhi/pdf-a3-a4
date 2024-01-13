package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/karmdip-mi/go-fitz"
	"github.com/nfnt/resize"
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

type offset struct {
	left int
	width int
}

func a3TOa4(fileName string, offsets []offset) []string{
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

	imgWidth := imc.Width
	number := len(offsets)
	halfWidth := imgWidth/number
	for i:= 0;i<number;i++ {
		newImg  := image.NewRGBA(image.Rect(0,0, halfWidth + offsets[i].width,imc.Height))
		left := halfWidth * i + offsets[i].left
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

	pageOffset2 := make(map[int][]offset)
	pageOffset2[0] = []offset{
		{left: 50},
		{left: 10,width: -200},
	}
	pageOffset2[1] = []offset{
		{left: 10},
		{left: 10,width: -200},
	}

	type pdfOffset map[int][]offset

	pdfPageOffset := make(map[string]pdfOffset)
	pdfPageOffset["青羊区2023期末真题.pdf"] = pageOffset2


	pageOffset1 := make(map[int][]offset)
	pageOffset1[0] = []offset{
		{left: 10,width: 150},
		{left: 150,width: -150},
		{left: 10,width: -150},
	}
	pageOffset1[1] = []offset{
		{left: 10,width: 150},
		{left: 150,width: -120},
		{left: 10,width: -100},
	}
	pdfPageOffset["金牛区2023期末真题.pdf"] = pageOffset1


	pageOffset := pdfPageOffset[in]
	for idx, imageName := range images {
		offsets := pageOffset[idx]
		fmt.Println("idx:",idx,"imageName:",imageName,"offsets:",offsets, "offsetsNum:",len(offsets))
		names := a3TOa4(imageName,  offsets)
		for _,name := range names {
			pdf.AddPage()
			body,_ := ioutil.ReadFile(name)
			img,_ ,err := image.Decode(bytes.NewReader(body))
			if err != nil {
				panic(err)
			}
			imc, _, _ := image.DecodeConfig(bytes.NewReader(body))
			fmt.Println("imgWidth:",imc.Width,"imgHeight:",imc.Height)
			_ = resize.Lanczos2
			m := resize.Resize(uint(rec.W*1.7), uint(rec.H*1.7), img, resize.Lanczos3)
			newName := name+".resized.jpg"
			f , err := os.Create(newName)
			if err != nil {
				panic(err)
			}
			err = jpeg.Encode(f,m,&jpeg.Options{
				Quality: 100,
			})
			if err != nil {
				panic(err)
			}
			pdf.Image(newName, 0, 0,nil)
		}
	}

	pdf.WritePdf(out)

}

