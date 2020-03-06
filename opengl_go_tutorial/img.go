package main

import (
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"image"
	"image/draw"
	_ "image/png"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
)

const (
	imgVertexShaderSource = `
   #version 410
   layout (location=0) in vec3 vp;
   layout (location=1) in vec2 texCoord;
   out vec2 tc;
   void main() {
       gl_Position = vec4(vp, 1.0);
       tc = texCoord;
   }
` + "\x00"

	imgFragmentShaderSource = `
   #version 410
   in vec2 tc;
   out vec4 frag_colour;
   uniform sampler2D samp;
   void main() {
       frag_colour = texture(samp, tc);
   }
` + "\x00"
)

type ImgDrawer struct {
	texID uint32
	file  string
}

func (i ImgDrawer) LoadTexture() error {
	imgFile, err := os.Open(i.file)
	if err != nil {
		return fmt.Errorf("texture %q not found on disk: %v", i.file, err)
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

	gl.GenTextures(1, &i.texID)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, i.texID)
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
	return nil
}

func (i ImgDrawer) LoadProgram(prog uint32) error {
	vertexShader, err := compileShader(imgVertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	fragmentShader, err := compileShader(imgFragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	return nil
}

func (i ImgDrawer) DrawScene(vao uint32, window *glfw.Window, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, i.texID)
	gl.BindVertexArray(vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(rectangleVertices)/3))

	glfw.PollEvents()
	window.SwapBuffers()
}
