package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/hysios/go-ffmpeg-player/player"
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

func setVideoTexture() {
	// video texture
	gl.GenTextures(1, &vidTexID)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	if vidTexID == 0 {
		panic("failed to create video texture")
	}
}

var (
	pixels = []float32{
		0.0, 0.0, 0.0, 1.0, 1.0, 1.0,
		1.0, 1.0, 1.0, 0.0, 0.0, 0.0,
	}
)

func drawVideo( /*vao uint32,*/ frame *player.Frame, window *glfw.Window, prog uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.BindTexture(gl.TEXTURE_2D, vidTexID)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(frame.Width), int32(frame.Height), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
	//gl.PixelStorei(gl.UNPACK_ROW_LENGTH, video.pitch/video.bpp)

	gl.UseProgram(prog)
	gl.Uniform2f(gl.GetUniformLocation(prog, gl.Str("TextureSize\x00")), float32(frame.Width), float32(frame.Height))
	gl.Uniform2f(gl.GetUniformLocation(prog, gl.Str("InputSize\x00")), float32(frame.Width), float32(frame.Height))
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, int32(frame.Width), int32(frame.Height), gl.RGB, gl.UNSIGNED_BYTE, nil)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, 2)

	//// draw image
	//gl.BindVertexArray(vao)
	//gl.DrawArrays(gl.TRIANGLES, 0, int32(len(rectangle)/3))

	glfw.PollEvents()
	window.SwapBuffers()
}
