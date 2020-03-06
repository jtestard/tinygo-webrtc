package main

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"os"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

const (
	vertexShaderSource = `
   #version 410
   layout (location=0) in vec3 vp;
   layout (location=1) in vec2 texCoord;
   out vec2 tc;
   void main() {
       gl_Position = vec4(vp, 1.0);
       tc = texCoord;
   }
` + "\x00"

	fragmentShaderSource = `
   #version 410
   in vec2 tc;
   out vec4 frag_colour;
   uniform sampler2D samp;
   void main() {
       frag_colour = texture(samp, tc);
   }
` + "\x00"
)

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func newImageTexture(file string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println(dir)
	imgFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("texture %q not found on disk: %v", file, err)
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return fmt.Errorf("could not decode: %w", err)
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))
	picTexID = texture
	return nil
}

// textures are stored "upside-down" in open gl
func texInvertY(texCoords []float32) {
	if len(texCoords)%2 != 0 || len(texCoords) < 2 {
		return
	}
	for i := 1; i < len(texCoords); i += 2 {
		texCoords[i] = 1 - texCoords[i]
	}
}
