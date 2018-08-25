# TODO

Merge the `Makefile` and the `build` script...

Make the '_test' and '_bin' directories part of the git repo... (Or really, learn to do a proper release...)

Example:

```
> image2rect --help

Usage of image2rect:
  -animation-build string
    	write the corresponding animated-build GIF file
  -animation-fps float
    	approximate animation frames per sec, 0.1-100 (default 15)
  -animation-pixels string
    	write the corresponding animated-pixels GIF file
  -center
    	center the output on zero (default true)
  -grayscale string
    	write the corresponding grayscale PNG file
  -input string
    	the input PNG, GIF, or JPEG file, default is stdin
  -invert
    	invert the image colors prior to grayscaling
  -monochrome string
    	write the corresponding monochrome PNG file
  -negative string
    	write the corresponding negative RGBA PNG file
  -output string
    	the output rectangle-data filename, default is stdout
  -report
    	report the pixel compression ratio to stderr
  -svg string
    	write the corresponding standalone SVG file
  -threshold uint
    	post-negation gray threshold, 0-255 (default 127)
  -verify string
    	write a verification RGBA color PNG file

```
