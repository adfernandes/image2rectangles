// GOPATH="$(pwd)" go build rect.go

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"os"
	"strings"
)

const style string = "fill: rgb(255,255,255); stroke: rgb(0,0,0); stroke-width: 0.03125;"

func getReaderFor(filename string) io.Reader {

	if filename == "-" || filename == "" {
		return os.Stdin
	}

	reader, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	return reader

}

func getWriterFor(filename string) io.Writer {

	if filename == "-" || filename == "" {
		return os.Stdout
	}

	writer, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	return writer

}

type quad struct {
	rects [4]image.Rectangle
	count [4]struct {
		expected int
		observed int
	}
	quads [4]*quad
}

func newQuad(gray *image.Gray, rect *image.Rectangle) *quad {

	min := rect.Min
	max := rect.Max

	mid := image.Point{(min.X + max.X) / 2, (min.Y + max.Y) / 2}

	q := new(quad)

	q.rects = [4]image.Rectangle{
		image.Rect(min.X, min.Y, mid.X, mid.Y),
		image.Rect(mid.X, min.Y, max.X, mid.Y),
		image.Rect(min.X, mid.Y, mid.X, max.Y),
		image.Rect(mid.X, mid.Y, max.X, max.Y),
	}

	for i, rect := range q.rects {

		q.count[i].expected = rect.Dx() * rect.Dy()

		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			for x := rect.Min.X; x < rect.Max.X; x++ {
				if gray.GrayAt(x, y).Y > 0 {
					q.count[i].observed++
				}
			}
		}

	}

	for i, rect := range q.rects {
		if q.count[i].expected > 1 && q.count[i].observed < q.count[i].expected {
			q.quads[i] = newQuad(gray, &rect)
		}
	}

	return q
}

func (q *quad) writeSvgRectangles(s *strings.Builder) {

	for i, rect := range q.rects {

		if q.count[i].expected > 0 {
			if q.count[i].observed == q.count[i].expected {
				s.WriteString(fmt.Sprintf("  <rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" style=\"%v\"/>\n", rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy(), style))
			}
		}

	}

	for _, sq := range q.quads {
		if sq != nil {
			sq.writeSvgRectangles(s)
		}
	}

}

func (q *quad) toSVG(rect *image.Rectangle) string {

	var s strings.Builder

	s.WriteString("<?xml version=\"1.0\" standalone=\"no\"?>\n")
	s.WriteString(fmt.Sprintf("<svg xmlns=\"http://www.w3.org/2000/svg\" version=\"1.1\" x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\">\n", rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy()))
	q.writeSvgRectangles(&s)
	s.WriteString("</svg>\n")

	return s.String()

}

func main() {

	var inputFilename string
	var outputFilename string
	var threshold uint
	var negate bool

	flag.StringVar(&inputFilename, "i", "-", "input filename, '-' for 'stdin'")
	flag.StringVar(&outputFilename, "o", "-", "output filename, '-' for 'stdout'")
	flag.UintVar(&threshold, "t", 127, "monochrome threshold, post negation")
	flag.BoolVar(&negate, "n", false, "negate the image colors prior to grayscaling")

	flag.Parse()

	input, format, err := image.Decode(getReaderFor(inputFilename))
	if err != nil {
		log.Fatal(err)
	}

	// we will assume that 'white' is the color that we want drawn

	fmt.Printf("read '%v' as a '%v' image with '%T' and bounds '%v'\n", inputFilename, format, input.ColorModel().Convert(color.RGBA{}), input.Bounds())

	if input.Bounds().Dx() < 2 || input.Bounds().Dy() < 2 {
		log.Fatal("the input image must be greater than 2 pixels wide and high")
	}

	if threshold < 0 || threshold > 255 {
		log.Fatal("the monochrome threshold myst be [0, 255] inclusive")
	}

	// convert the input image to the RGBA (premultiplied alpha) color space

	rgba := image.NewRGBA(input.Bounds())

	for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y++ {
		for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x++ {
			rgba.Set(x, y, input.At(x, y))
		}
	}

	// negate the image colors, if requested, respecting the alpha channel

	if negate {
		for y := input.Bounds().Min.Y; y < input.Bounds().Max.Y; y++ {
			for x := input.Bounds().Min.X; x < input.Bounds().Max.X; x++ {
				r, g, b, a := rgba.At(x, y).RGBA()
				r = a - r
				g = a - g
				b = a - b
				rgba.Set(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
			}
		}
	}

	// convert the image to grayscale, removing the alpha channel

	gray := image.NewGray(rgba.Bounds())

	for y := rgba.Bounds().Min.Y; y < rgba.Bounds().Max.Y; y++ {
		for x := rgba.Bounds().Min.X; x < rgba.Bounds().Max.X; x++ {
			gray.Set(x, y, rgba.At(x, y))
		}
	}

	// now convert to monochrome

	black := color.Gray{0}
	white := color.Gray{255}
	mono := image.NewGray(gray.Bounds())

	for y := gray.Bounds().Min.Y; y < gray.Bounds().Max.Y; y++ {
		for x := gray.Bounds().Min.X; x < gray.Bounds().Max.X; x++ {
			if uint(gray.GrayAt(x, y).Y) > threshold {
				mono.SetGray(x, y, white)
			} else {
				mono.SetGray(x, y, black)
			}
		}
	}

	// save the covnerted image

	// TODO png.Encode(getWriterFor(outputFilename), gray)
	// TODO png.Encode(getWriterFor(outputFilename), mono)

	// write the debugging svg file

	rect := mono.Bounds()
	q := newQuad(mono, &rect)

	svg := q.toSVG(&rect)
	writer := getWriterFor(outputFilename)
	writer.Write([]byte(svg))

}
