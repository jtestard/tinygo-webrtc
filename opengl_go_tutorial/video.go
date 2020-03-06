package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
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
   in vec2 tc;
   out vec4 frag_colour;
   uniform sampler2D samp;
   void main() {
       frag_colour = texture(samp, tc);
   }
` + "\x00"
)

func playVideo(inputfile string, frames chan<- *player.Frame) {
	ply, _ := player.Open(inputfile, &player.Options{Loop: true})
	ply.SetScale(width, height)
	ply.Play()

	ply.PreFrame(func(frame *player.Frame) {
		frames <- frame
	})
	ply.Wait()
}

func newVideoTexture(file string) (uint32, chan *player.Frame, error) {
	chFrames := make(chan *player.Frame, 100)
	go playVideo(file, chFrames)

	firstFrame := <-chFrames
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
		gl.RGB,
		firstFrame.Width,
		firstFrame.Height,
		0,
		gl.SRGB,
		gl.UNSIGNED_BYTE,
		gl.Ptr(firstFrame.Data[0]))
	return 0, chFrames, nil
}

func setupVideoShaders(prog uint32) {
	vertexShader, err := compileShader(videoVertexShaderSource, gl.VERTEX_SHADER)
	checkNoError(err)

	fragmentShader, err := compileShader(videoFragmentShaderSource, gl.FRAGMENT_SHADER)
	checkNoError(err)
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
}
