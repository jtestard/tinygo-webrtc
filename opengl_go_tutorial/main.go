package main

import (
	"log"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/hysios/go-ffmpeg-player/player"
)

const (
	width  = 500
	height = 500
)

var (
	//triangle = []float32{
	//	0, 0.5, 0,
	//	-0.5, -0.5, 0,
	//	0.5, -0.5, 0,
	//}
	showVideo = true
	vidTexID  uint32
	picTexID  uint32
	rectangle = []float32{
		-1, -1, 0,
		-1, 1, 0,
		1, -1, 0, // left triangle
		-1, 1, 0,
		1, -1, 0,
		1, 1, 0, // right triangle
	}
)

func main() {
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()
	program := initOpenGL()

	if showVideo {
		chFrames := make(chan *player.Frame, 100)
		go playVideo("output.ivf", chFrames)

		for !window.ShouldClose() {
			drawVideo(<-chFrames, window, program)
		}
	} else {
		vao := makeVao(rectangle)
		for !window.ShouldClose() {
			draw(vao, window, program)
		}
	}
}

func draw(vao uint32, window *glfw.Window, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	gl.BindVertexArray(vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(rectangle)/3))

	glfw.PollEvents()
	window.SwapBuffers()
}

// initGlfw initializes glfw and returns a Window to use.
func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "Play Video", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

// initOpenGL initializes OpenGL and returns an intiialized program.
func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	prog := gl.CreateProgram()

	//gl.GenTextures(1, &picTexID)
	//gl.BindTexture(gl.TEXTURE_2D, picTexID)
	//gl.TexParameteri(gl.GLTex)
	if showVideo {
		setVideoTexture()
	} else {
		vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
		if err != nil {
			panic(err)
		}

		fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
		if err != nil {
			panic(err)
		}
		gl.AttachShader(prog, vertexShader)
		gl.AttachShader(prog, fragmentShader)
	}
	gl.LinkProgram(prog)
	return prog
}

// makeVao initializes and returns a vertex array from the points provided.
func makeVao(points []float32) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	return vao
}
