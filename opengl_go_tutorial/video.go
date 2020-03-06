package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/hysios/go-ffmpeg-player/player"
)

const (
	videoVertexShaderSource = `
   #version 410
   layout (location=0) in vec3 vp;
   layout (location=1) in vec2 texCoord;
   out vec2 tc;
   void main() {
       gl_Position = vec4(vp, 1.0);
       tc = texCoord;
   }
` + "\x00"

	videoFragmentShaderSource = `
   #version 410
   in vec2 tcY;
   out vec4 frag_colour;
   uniform sampler2D sampY;
   uniform sampler2D sampU;
   uniform sampler2D sampV;
   void main() {
       frag_colour = texture(samp, tc);
   }
` + "\x00"
)

type VideoDrawer struct {
	texID    uint32
	file     string
	chFrames chan *player.Frame
}

func (v *VideoDrawer) LoadTexture(prog uint32) error {
	v.chFrames = make(chan *player.Frame, 100)
	go v.playVideo()

	gl.GenTextures(1, &v.texID)
	v.loadNextTexture(prog)
	return nil
}

func fromGoString(msg string) *uint8 {
	out := make([]uint8, len(msg))
	for i, b := range msg {
		out[i] = uint8(b)
	}
	return &out[0]
}

func (v *VideoDrawer) loadNextTexture(prog uint32) {
	frame := <-v.chFrames

	gl.ActiveTexture(gl.TEXTURE1)
	i := gl.GetUniformLocation(prog, fromGoString("sampU")) // Get reference to variable sampU from program
	gl.Uniform1i(i, 1)                                      // Binds SampU variable from GLSL to texture unit
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.BindTexture(gl.TEXTURE_2D, 1)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGB,
		frame.Width,
		frame.Height,
		0,
		gl.SRGB,
		gl.UNSIGNED_BYTE,
		gl.Ptr(frame.Data[1]))

	// 1
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, v.texID)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGB,
		frame.Width,
		frame.Height,
		0,
		gl.SRGB,
		gl.UNSIGNED_BYTE,
		gl.Ptr(frame.Data[0]))
}

func (v *VideoDrawer) LoadProgram(prog uint32) error {
	vertexShader, err := compileShader(videoVertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	fragmentShader, err := compileShader(videoFragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	return nil
}

func (v *VideoDrawer) DrawScene(vao uint32, window *glfw.Window, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, v.texID)
	gl.BindVertexArray(vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(rectangleVertices)/3))

	glfw.PollEvents()
	window.SwapBuffers()
}

func (v *VideoDrawer) playVideo() {
	ply, _ := player.Open(v.file, &player.Options{Loop: true})
	ply.SetScale(width, height)
	ply.Play()

	ply.PreFrame(func(frame *player.Frame) {
		v.chFrames <- frame
	})
	ply.Wait()
}
