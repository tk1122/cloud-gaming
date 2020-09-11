package ui

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"

	"github.com/fogleman/nes/nes"
)

var homeDir string

func init() {
	u, err := user.Current()
	if err != nil {
		log.Fatalln(err)
	}
	homeDir = u.HomeDir
}

func thumbnailURL(hash string) string {
	return "http://www.michaelfogleman.com/static/nes/" + hash + ".png"
}

func thumbnailPath(hash string) string {
	return homeDir + "/.nes/thumbnail/" + hash + ".png"
}

func sramPath(hash string) string {
	return homeDir + "/.nes/sram/" + hash + ".dat"
}

func savePath(hash string) string {
	return homeDir + "/.nes/save/" + hash + ".dat"
}

func combineButtons(a, b [8]bool) [8]bool {
	var result [8]bool
	for i := 0; i < 8; i++ {
		result[i] = a[i] || b[i]
	}
	return result
}

func hashFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(data)), nil
}

func copyImage(src image.Image) *image.RGBA {
	dst := image.NewRGBA(src.Bounds())
	draw.Draw(dst, dst.Rect, src, image.ZP, draw.Src)
	return dst
}

func loadPNG(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return png.Decode(file)
}

func savePNG(path string, im image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, im)
}

func saveGIF(path string, frames []image.Image) error {
	var palette []color.Color
	for _, c := range nes.Palette {
		palette = append(palette, c)
	}
	g := gif.GIF{}
	for i, src := range frames {
		if i%3 != 0 {
			continue
		}
		dst := image.NewPaletted(src.Bounds(), palette)
		draw.Draw(dst, dst.Rect, src, image.ZP, draw.Src)
		g.Image = append(g.Image, dst)
		g.Delay = append(g.Delay, 5)
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return gif.EncodeAll(file, &g)
}

func writeSRAM(filename string, sram []byte) error {
	dir, _ := path.Split(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return binary.Write(file, binary.LittleEndian, sram)
}

func readSRAM(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	sram := make([]byte, 0x2000)
	if err := binary.Read(file, binary.LittleEndian, sram); err != nil {
		return nil, err
	}
	return sram, nil
}
