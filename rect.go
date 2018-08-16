// GOPATH="$(pwd)" go build rect.go

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	_ "image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"strings"
)

const style string = "fill: rgb(255,255,255); stroke: rgb(0,0,0); stroke-width: 0.03125;"

func getDefaultReaderFor(filename string) *os.File {

	if filename == "-" || filename == "" {
		return os.Stdin
	}

	reader, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	return reader

}

func getDefaultWriterFor(filename string) *os.File {

	if filename == "-" || filename == "" {
		return os.Stdout
	}

	writer, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	return writer

}

func getWriterFor(filename string) *os.File {

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

	log.SetFlags(log.Llongfile)

	var inputFilename string
	var outputFilename string

	var colorFilename string

	var invert bool
	var negativeFilename string

	var grayThreshold uint
	var grayFilename string

	var monoFilename string
	var animationFilename string
	var animationFramesPerSecond float64
	var animationDelay int

	var svgFilename string

	flag.StringVar(&inputFilename, "input", "", "the input PNG, GIF, or JPEG file, default is stdin")
	flag.StringVar(&outputFilename, "output", "", "the output rectangle-data filename, default is stdout")

	flag.StringVar(&colorFilename, "verify", "", "write a verification RGBA color PNG file")

	flag.BoolVar(&invert, "invert", false, "invert the image colors prior to grayscaling")
	flag.StringVar(&negativeFilename, "negative", "", "write the corresponding negative RGBA PNG file")

	flag.UintVar(&grayThreshold, "threshold", 127, "monochrome gray threshold, post negation, 0-255")
	flag.StringVar(&grayFilename, "grayscale", "", "write the corresponding grayscale PNG file")

	flag.StringVar(&monoFilename, "monochrome", "", "write the corresponding monochrome PNG file")
	flag.StringVar(&animationFilename, "animation", "", "write the corresponding animated GIF file")
	flag.Float64Var(&animationFramesPerSecond, "animation-fps", 5.0, "approximate animation frames per sec, 0.1-100")

	flag.StringVar(&svgFilename, "svg", "", "write the corresponding standalone SVG file")

	flag.Parse()

	if !invert && negativeFilename != "" {
		log.Fatal("a negative image was requested, but color inversion was not")
	}

	if grayThreshold < 0 || grayThreshold > 255 {
		log.Fatal("the monochrome threshold myst be in [0, 255] inclusive")
	}

	if animationFramesPerSecond < 0.1 || animationFramesPerSecond > 100 {
		log.Fatal("the requested animation FPS is out of range")
	}

	animationDelay = int(math.Round(100.0 / animationFramesPerSecond))

	inputReader := getDefaultReaderFor(inputFilename)
	input, format, err := image.Decode(inputReader)
	if err != nil {
		log.Fatal(err)
	}
	inputReader.Close()

	// we will assume that 'white' is the color that we want drawn

	bounds := input.Bounds()
	fmt.Printf("read '%v' as a '%v' image with '%T' and bounds '%v'\n", inputFilename, format, input.ColorModel().Convert(color.RGBA{}), bounds)

	if bounds.Dx() < 2 || bounds.Dy() < 2 {
		log.Fatal("the input image must be greater than 2 pixels wide and high")
	}

	if bounds.Min.X != 0 || bounds.Min.Y != 0 {
		log.Fatal("internal error: the image 'Min' bound must be (0,0) but it isn't")
	}

	// convert the input image to the RGBA (premultiplied alpha) color space

	rgba := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, input.At(x, y))
		}
	}

	if colorFilename != "" {
		writer := getWriterFor(colorFilename)
		err = png.Encode(writer, rgba)
		if err != nil {
			log.Fatal(err)
		}
		writer.Close()
	}

	// negate the image colors, if requested, respecting the alpha channel

	if invert {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, a := rgba.At(x, y).RGBA()
				r = a - r
				g = a - g
				b = a - b
				rgba.Set(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
			}
		}
	}

	if negativeFilename != "" {
		writer := getWriterFor(negativeFilename)
		err = png.Encode(writer, rgba)
		if err != nil {
			log.Fatal(err)
		}
		writer.Close()
	}

	// convert the image to grayscale, removing the alpha channel

	gray := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray.Set(x, y, rgba.At(x, y))
		}
	}

	if grayFilename != "" {
		writer := getWriterFor(grayFilename)
		err = png.Encode(writer, gray)
		if err != nil {
			log.Fatal(err)
		}
		writer.Close()
	}

	// now convert to monochrome

	black := color.Gray{0}
	white := color.Gray{255}
	mono := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if uint(gray.GrayAt(x, y).Y) > grayThreshold {
				mono.SetGray(x, y, white)
			} else {
				mono.SetGray(x, y, black)
			}
		}
	}

	if monoFilename != "" {
		writer := getWriterFor(monoFilename)
		err = png.Encode(writer, mono)
		if err != nil {
			log.Fatal(err)
		}
		writer.Close()
	}

	// FIXME write the debugging svg file

	rect := mono.Bounds()
	q := newQuad(mono, &rect)

	svg := q.toSVG(&rect)
	writer := getWriterFor(svgFilename)
	_, err = writer.Write([]byte(svg))
	if err != nil {
		log.Fatal(err)
	}
	writer.Close()

	// FIXME test the maximal area thingie AND SANVE THE SVG FILE

	animation := &gif.GIF{}
	gifOptions := gif.Options{NumColors: 2}

	for {

		// convert the 'mono' image to a single GIF frame

		var byteBuffer bytes.Buffer
		writer := bufio.NewWriter(&byteBuffer)
		err = gif.Encode(writer, mono, &gifOptions)
		if err != nil {
			log.Fatal(err)
		}

		// re-read that single GIF frame back into an Image

		reader := bytes.NewReader(byteBuffer.Bytes())
		frame, err := gif.Decode(reader)
		if err != nil {
			log.Fatal(err)
		}

		// append the GIF frame onto the animation array

		animation.Image = append(animation.Image, frame.(*image.Paletted))
		animation.Delay = append(animation.Delay, animationDelay)

		// *** png.Encode(getWriterFor(fmt.Sprintf("_zzz/%s~%03d.png", outputFilename, count)), mono)

		area, rect := maximalRectangle(mono)

		// *** fmt.Printf("count: %v, area: %v, rect: %v\n", count, area, rect)

		for y := rect.Bounds().Min.Y; y < rect.Bounds().Max.Y; y++ {
			for x := rect.Bounds().Min.X; x < rect.Bounds().Max.X; x++ {
				mono.SetGray(x, y, black)
			}
		}

		if area <= 0 {
			break
		}

	}

	// *** START HERE create the animated GIF and write (and close) it

}

// https://stackoverflow.com/a/20039017
//
// We assume that we have already checked that the input
// rectangle has a minimum x and y bound value of zero.
//
func maximalRectangle(img *image.Gray) (int, image.Rectangle) {

	M := img.Bounds().Max.X // length of a row
	N := img.Bounds().Max.Y // number of rows

	type Pair struct {
		one int
		two int
	}

	bestLl := Pair{0, 0}
	bestUr := Pair{-1, -1}
	var bestArea int

	c := make([]int, M+1)  // cache
	s := make([]Pair, M+1) // stack
	var top int            // top of stack
	var row int            // cache row

	// main algorithm

	for n := 0; n != N; n++ {

		var openWidth int

		// update cache

		for m := 0; m != M; m++ {
			b := img.GrayAt(row, m).Y
			if b == 0 {
				c[m] = 0
			} else {
				c[m]++
			}
		}
		row++

		for m := 0; m != M+1; m++ {

			if c[m] > openWidth { // open a new rectangle?

				// push(m, openWidth)
				s[top].one = m
				s[top].two = openWidth
				top++

				openWidth = c[m]

			} else if c[m] < openWidth { // close rectangle(s)?

				var m0, w0, area int

				for {

					// pop(&m0, &w0)
					top--
					m0 = s[top].one
					w0 = s[top].two

					area = openWidth * (m - m0)

					if area > bestArea {
						bestArea = area
						bestLl.one = m0
						bestLl.two = n
						bestUr.one = m - 1
						bestUr.two = n - openWidth + 1
					}

					openWidth = w0

					if c[m] >= openWidth {
						break
					}

				}

				openWidth = c[m]

				if openWidth != 0 {

					// push(m0, w0)
					s[top].one = m0
					s[top].two = w0
					top++

				}

			}

		}

	}

	if bestArea > 0 {
		fmt.Printf("image.Rect(%v, %v, %v, %v)\n", bestLl.two+1, bestUr.one+1, bestUr.two, bestLl.one)
		return bestArea, image.Rect(bestLl.two+1, bestUr.one+1, bestUr.two, bestLl.one)
	}

	return 0, image.Rect(0, 0, 0, 0)

}
