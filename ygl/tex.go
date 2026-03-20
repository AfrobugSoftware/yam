package ygl

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	gl "github.com/chsc/gogl/gl33"
)

func CreateTex2D(filePath string, minFilter, maxFilter gl.Enum) (gl.Uint, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open image file: %v", err)
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return 0, fmt.Errorf("failed to decode image: %v", err)
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	var texId gl.Uint
	gl.GenTextures(1, &texId)
	gl.BindTexture(gl.TEXTURE_2D, texId)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.Int(minFilter))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.Int(maxFilter))

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		gl.Sizei(bounds.Max.X),
		gl.Sizei(bounds.Max.Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Pointer(&rgba.Pix[0]),
	)
	return texId, nil
}

func SetActiveTex(tex gl.Uint) {
	gl.BindTexture(gl.TEXTURE_2D, tex)
}

func DestroyTex2D(tex gl.Uint) {
	gl.DeleteTextures(1, &tex)
}
