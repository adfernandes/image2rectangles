// GOPATH="$(pwd)" go build rect.go

package main

import (
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

type rect struct {
	x, y, dx, dy float32
}

type rectImage struct {
	bounds rect
	pixels []rect
}

func newRect(rectangle *image.Rectangle) rect {
	r := rect{
		x:  float32(rectangle.Min.X),
		y:  float32(rectangle.Min.Y),
		dx: float32(rectangle.Dx()),
		dy: float32(rectangle.Dy()),
	}
	return r
}

func newRectImage(rectangle *image.Rectangle) rectImage {
	ri := rectImage{bounds: newRect(rectangle)}
	return ri
}

func (ri *rectImage) Center() {
	ox := ri.bounds.x + 0.5*ri.bounds.dx
	oy := ri.bounds.y + 0.5*ri.bounds.dy
	ri.bounds.x -= ox
	ri.bounds.y -= oy
	for i := range ri.pixels {
		ri.pixels[i].x -= ox
		ri.pixels[i].y -= oy
	}
}

func (ri *rectImage) String() string {

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("box %v %v %v %v\n", ri.bounds.x, ri.bounds.y, ri.bounds.dx, ri.bounds.dy))
	sb.WriteString(fmt.Sprintf("  pixels %v\n", len(ri.pixels)))
	for _, r := range ri.pixels {
		sb.WriteString(fmt.Sprintf("    rect %v %v %v %v\n", r.x, r.y, r.dx, r.dy))
	}

	return sb.String()

}

func (ri *rectImage) toSVG() string {

	const strokeWidth float32 = 0.03
	const offset float32 = strokeWidth / 2.0

	var sb strings.Builder

	sb.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"no\"?>\n")
	sb.WriteString(fmt.Sprintf("<!DOCTYPE svg PUBLIC \"-//W3C//DTD SVG 1.1//EN\" \"http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd\">\n"))

	sb.WriteString(fmt.Sprintf("<svg xmlns=\"http://www.w3.org/2000/svg\" version=\"1.1\"\n"))
	sb.WriteString(fmt.Sprintf("     width=\"100%%\" height=\"100%%\"\n"))
	sb.WriteString(fmt.Sprintf("     preserveAspectRatio=\"xMidYMid meet\"\n"))
	sb.WriteString(fmt.Sprintf("     viewBox=\"%v %v %v %v\">\n", ri.bounds.x-offset, ri.bounds.y-offset, ri.bounds.dx+offset, ri.bounds.dy+offset))

	sb.WriteString(fmt.Sprintf("    <g fill=\"gray\" stroke=\"none\">\n"))
	sb.WriteString(fmt.Sprintf("        <rect x=\"%v\" y=\"%v\" width=\"%v\" height=\"%v\"/>\n", ri.bounds.x, ri.bounds.y, ri.bounds.dx, ri.bounds.dy))

	sb.WriteString(fmt.Sprintf("        <g fill=\"white\" stroke=\"black\" stroke-width=\"%v\">\n", strokeWidth))
	for _, r := range ri.pixels {
		sb.WriteString(fmt.Sprintf("            <rect x=\"%v\" y=\"%v\" width=\"%v\" height=\"%v\"/>\n", r.x, r.y, r.dx, r.dy))
	}
	sb.WriteString(fmt.Sprintf("        </g>\n"))

	sb.WriteString(fmt.Sprintf("    </g>\n"))
	sb.WriteString(fmt.Sprintf("</svg>\n"))

	return sb.String()

}

func main() {

	log.SetFlags(log.Llongfile)

	var inputFilename string
	var outputFilename string
	var centerOutput bool

	var colorFilename string

	var invert bool
	var negativeFilename string

	var grayThreshold uint
	var grayFilename string

	var monoFilename string
	var animationBuildFilename string
	var animationPixelsFilename string
	var animationRequested bool
	var animationFramesPerSecond float64
	var animationDelay int

	var svgFilename string

	var report bool

	flag.StringVar(&inputFilename, "input", "", "the input PNG, GIF, or JPEG file, default is stdin")
	flag.StringVar(&outputFilename, "output", "", "the output rectangle-data filename, default is stdout")
	flag.BoolVar(&centerOutput, "center", true, "center the output on zero")

	flag.StringVar(&colorFilename, "verify", "", "write a verification RGBA color PNG file")

	flag.BoolVar(&invert, "invert", false, "invert the image colors prior to grayscaling")
	flag.StringVar(&negativeFilename, "negative", "", "write the corresponding negative RGBA PNG file")

	flag.UintVar(&grayThreshold, "threshold", 127, "monochrome gray threshold, post negation, 0-255")
	flag.StringVar(&grayFilename, "grayscale", "", "write the corresponding grayscale PNG file")

	flag.StringVar(&monoFilename, "monochrome", "", "write the corresponding monochrome PNG file")
	flag.StringVar(&animationBuildFilename, "animation-build", "", "write the corresponding animated-build GIF file")
	flag.StringVar(&animationPixelsFilename, "animation-pixels", "", "write the corresponding animated-pixels GIF file")
	flag.Float64Var(&animationFramesPerSecond, "animation-fps", 5.0, "approximate animation frames per sec, 0.1-100")

	flag.StringVar(&svgFilename, "svg", "", "write the corresponding standalone SVG file")

	flag.BoolVar(&report, "report", false, "report the pixel compression ratio to stderr")

	flag.Parse()

	if !invert && negativeFilename != "" {
		log.Fatal("a negative image was requested, but color inversion was not")
	}

	if grayThreshold < 0 || grayThreshold > 255 {
		log.Fatal("the monochrome threshold myst be in [0, 255] inclusive")
	}

	if animationBuildFilename != "" || animationPixelsFilename != "" {
		animationRequested = true
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

	if false {
		fmt.Printf("read '%v' as a '%v' image with '%T' and bounds '%v'\n", inputFilename, format, input.ColorModel().Convert(color.RGBA{}), bounds)
	}

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

	// decompose the image into maximal rectangles, capturing the image if requested

	animation := &gif.GIF{}

	palette := []color.Color{color.Transparent, color.White}

	const gifDisposalUnspecified = 0
	const gifDisposalDoNotDispose = 1
	const gifDisposalRestoreToBackgroundColor = 2
	const gifDisposalRestoreToPrevious = 3

	if animationRequested {

		animation.Image = append(animation.Image, image.NewPaletted(bounds, palette))
		animation.Delay = append(animation.Delay, animationDelay)
		animation.Disposal = append(animation.Disposal, gifDisposalUnspecified)
	}

	var pixels int
	boxen := newRectImage(&bounds)

	for {

		area, rect := maximalRectangle(mono)

		for y := rect.Bounds().Min.Y; y < rect.Bounds().Max.Y; y++ {
			for x := rect.Bounds().Min.X; x < rect.Bounds().Max.X; x++ {
				mono.SetGray(x, y, black)
			}
		}

		if area <= 0 {
			break
		}

		boxen.pixels = append(boxen.pixels, newRect(&rect))
		pixels += rect.Dx() * rect.Dy()

		if animationRequested {

			// convert the 'mono' image to a single GIF frame

			frame := image.NewPaletted(bounds, palette)

			for y := rect.Bounds().Min.Y; y < rect.Bounds().Max.Y; y++ {
				for x := rect.Bounds().Min.X; x < rect.Bounds().Max.X; x++ {
					frame.SetColorIndex(x, y, 1)
				}
			}

			// append the GIF frame onto the animation array

			animation.Image = append(animation.Image, frame)
			animation.Delay = append(animation.Delay, animationDelay)
			animation.Disposal = append(animation.Disposal, gifDisposalUnspecified)

		}

	}

	if animationBuildFilename != "" {

		for i := range animation.Delay {
			animation.Disposal[i] = gifDisposalDoNotDispose
		}

		writer := getWriterFor(animationBuildFilename)
		err = gif.EncodeAll(writer, animation)
		if err != nil {
			log.Fatal(err)
		}
		writer.Close()
	}

	if animationPixelsFilename != "" {

		for i := range animation.Delay {
			animation.Disposal[i] = gifDisposalRestoreToBackgroundColor
		}

		writer := getWriterFor(animationPixelsFilename)
		err = gif.EncodeAll(writer, animation)
		if err != nil {
			log.Fatal(err)
		}
		writer.Close()
	}

	if centerOutput {
		boxen.Center()
	}

	if svgFilename != "" {

		writer := getWriterFor(svgFilename)
		_, err = writer.WriteString(boxen.toSVG())
		if err != nil {
			log.Fatal(err)
		}
		writer.Close()

	}

	writer := getDefaultWriterFor(outputFilename)
	writer.Write([]byte(boxen.String()))
	writer.Close()

	// report the compression ratio, if requested

	if report {
		fmt.Fprintf(os.Stderr, "pixels: { in: %v, out: %v }\n", pixels, len(boxen.pixels))
	}

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
		return bestArea, image.Rect(bestLl.two+1, bestUr.one+1, bestUr.two, bestLl.one)
	}

	return 0, image.ZR

}
