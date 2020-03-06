package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	// "github.com/hysios/go-ffmpeg-player/player"
	"github.com/pkg/errors"
)

const (
	width  = 500
	height = 500
)

var (
	rectangleVertices = []float32{
		-1, -1, 0, // A
		-1, 1, 0, // B
		1, -1, 0, // C left triangle
		-1, 1, 0, // B
		1, -1, 0, // C
		1, 1, 0, // D right triangle
	}
	rectangleTexCoords = []float32{
		0, 0, // A
		0, 1, // B
		1, 0, // C
		0, 1, // B
		1, 0, // C
		1, 1, // D
	}
)

func main() {
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()
	drawer, err := NewDrawer("profile.png")
	checkNoError(err)
	program := initOpenGL(drawer)
	vao := makeVao(rectangleVertices, rectangleTexCoords)
	err = drawer.LoadTexture()
	checkNoError(err)
	for !window.ShouldClose() {
		drawer.DrawScene(vao, window, program)
	}
}

func checkNoError(err error) {
	if err != nil {
		panic(fmt.Sprintf("%v", errors.WithStack(err)))
	}
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
	checkNoError(err)
	window.MakeContextCurrent()

	return window
}

// initOpenGL initializes OpenGL and returns an intiialized program.
func initOpenGL(drawer Drawer) uint32 {
	err := gl.Init()
	checkNoError(err)
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	prog := gl.CreateProgram()
	drawer.LoadProgram(prog)
	gl.LinkProgram(prog)
	return prog
}

// makeVao initializes and returns a vertex array from the points provided.
func makeVao(vertices []float32, textureCoords []float32) uint32 {
	vbos := make([]uint32, 2)
	// vertices
	gl.GenBuffers(1, &vbos[0])
	gl.BindBuffer(gl.ARRAY_BUFFER, vbos[0])
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(vertices), gl.Ptr(vertices), gl.STATIC_DRAW)

	// texture coords
	texInvertY(textureCoords)
	gl.GenBuffers(1, &vbos[1])
	gl.BindBuffer(gl.ARRAY_BUFFER, vbos[1])
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(textureCoords), gl.Ptr(textureCoords), gl.STATIC_DRAW)

	// create vao
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	// bind vertices
	gl.BindBuffer(gl.ARRAY_BUFFER, vbos[0])
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)
	gl.EnableVertexAttribArray(0)

	// bind textures
	gl.BindBuffer(gl.ARRAY_BUFFER, vbos[1])
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 0, nil)
	gl.EnableVertexAttribArray(1)

	return vao
}
