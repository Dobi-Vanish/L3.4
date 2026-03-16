package processor

import (
	"image"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"image/gif"
)

type Processor struct{}

func NewProcessor() *Processor {
	return &Processor{}
}

func (p *Processor) Process(inputPath, outputDir, baseName string) (resizedPath, thumbPath, watermarkedPath string, err error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return "", "", "", err
	}
	defer file.Close()

	var src image.Image

	gifImg, err := gif.DecodeAll(file)
	if err == nil && len(gifImg.Image) > 0 {
		src = gifImg.Image[0]
	} else {
		file.Seek(0, 0)
		src, _, err = image.Decode(file)
		if err != nil {
			return "", "", "", err
		}
	}

	resized := imaging.Resize(src, 800, 0, imaging.Lanczos)
	resizedPath = filepath.Join(outputDir, baseName+"_resized.jpg")
	if err := imaging.Save(resized, resizedPath); err != nil {
		return "", "", "", err
	}

	thumb := imaging.Fill(src, 150, 150, imaging.Center, imaging.Lanczos)
	thumbPath = filepath.Join(outputDir, baseName+"_thumb.jpg")
	if err := imaging.Save(thumb, thumbPath); err != nil {
		return "", "", "", err
	}

	watermarked := p.addWatermark(src)
	watermarkedPath = filepath.Join(outputDir, baseName+"_watermarked.jpg")
	if err := imaging.Save(watermarked, watermarkedPath); err != nil {
		return "", "", "", err
	}

	return resizedPath, thumbPath, watermarkedPath, nil
}

func (p *Processor) addWatermark(img image.Image) image.Image {
	dst := imaging.Clone(img)
	bounds := dst.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	textImg := image.NewRGBA(image.Rect(0, 0, w, h))
	fontFace := basicfont.Face7x13
	drawer := &font.Drawer{
		Dst:  textImg,
		Src:  image.White,
		Face: fontFace,
		Dot:  fixed.P(w-100, h-20),
	}
	drawer.DrawString("SAMPLE")

	watermarked := imaging.Overlay(dst, textImg, image.Pt(0, 0), 0.5)
	return watermarked
}
