package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/libretro/ludo/libretro"
	"github.com/pion/webrtc/pkg/media"
	"github.com/pion/webrtc/pkg/media/ivfreader"
)

func init() {
	runtime.LockOSThread()
}

func checkNoError(err error) {
	if err != nil {
		panic(err)
	}
}

func checkNoErrorWithMsg(msg string, err error) {
	if err != nil {
		panic(fmt.Errorf(msg, err))
	}
}

var (
	window  *glfw.Window
	program uint32 // current program used for the game quad
	vao     uint32
	vbo     uint32
	texID   uint32
	fboID   uint32
	rboID   uint32

	pitch         int32  // pitch set by the refresh callback
	pixFmt        uint32 // format set by the environment callback
	pixType       uint32
	bpp           int32
	width, height int32 // dimensions set by the refresh
	// callback

	Geom libretro.GameGeometry
)

var vertices = []float32{
	//  X, Y, U, V
	-1.0, -1.0, 0.0, 1.0, // left-bottom
	-1.0, 1.0, 0.0, 0.0, // left-top
	1.0, -1.0, 1.0, 1.0, // right-bottom
	1.0, 1.0, 1.0, 0.0, // right-top
}

func main() {
	configure(true)

	// Open a IVF file and start reading using our IVFReader
	file, ivfErr := os.Open("output.ivf")
	checkNoError(ivfErr)

	ivf, header, ivfErr := ivfreader.NewWith(file)
	checkNoError(ivfErr)

	// Send our video file frame at a time. Pace our sending so we send it at the same speed it should be played back as.
	// This isn't required since the video is timestamped, but we will such much higher loss if we send all at once.
	sleepTime := time.Millisecond * time.Duration((float32(header.TimebaseNumerator)/float32(header.TimebaseDenominator))*1000)
	for !window.ShouldClose() {
		frame, _, ivfErr := ivf.ParseNextFrame()
		if ivfErr == io.EOF {
			break
		}
		checkNoError(ivfErr)

		time.Sleep(sleepTime)
		ivfErr = videoTrack.WriteSample(media.Sample{Data: frame, Samples: 90000})
		checkNoError(ivfErr)

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		window.SwapBuffers()
		glfw.PollEvents()
	}
	fmt.Println("video completed")
}

func configure(fullscreen bool) {
	err := glfw.Init()
	checkNoErrorWithMsg("could not initialize glfw: %v", err)

	var m *glfw.Monitor

	if fullscreen {
		m = glfw.GetMonitors()[0]
		vm := m.GetVideoMode()
		width = int32(vm.Width)
		height = int32(vm.Height)
	} else {
		width = 320 * 3
		height = 180 * 3
	}

	// On OSX we have to force a core profile to not end up with 2.1 which cause
	// a font drawing issue
	if runtime.GOOS == "darwin" {
		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 2)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	} else {
		glfw.WindowHint(glfw.ContextVersionMajor, 2)
		glfw.WindowHint(glfw.ContextVersionMinor, 1)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLAnyProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.False)
	}

	window, err = glfw.CreateWindow(800, 600, "Hello world", nil, nil)
	checkNoErrorWithMsg("could not create opengl renderer: %v", err)

	window.MakeContextCurrent()

	// Force a minimum size for the window.
	window.SetSizeLimits(160, 120, glfw.DontCare, glfw.DontCare)
	window.SetInputMode(glfw.CursorMode, glfw.CursorHidden)

	err = gl.Init()
	checkNoError(err)

	gl.ClearColor(0, 0.5, 1.0, 1.0)

	fbw, fbh := window.GetFramebufferSize()

	// No shaders
	// No update of filter

	textureUniform := gl.GetUniformLocation(program, gl.Str("Texture\x00"))
	gl.Uniform1i(textureUniform, 0)

	// Configure the vertex data
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))

	texCoordAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vertTexCoord\x00")))
	gl.EnableVertexAttribArray(texCoordAttrib)
	gl.VertexAttribPointer(texCoordAttrib, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))

	// Some cores won't call SetPixelFormat, provide default values
	if pixFmt == 0 {
		pixFmt = gl.UNSIGNED_SHORT_5_5_5_1
		pixType = gl.BGRA
		bpp = 2
	}

	gl.GenTextures(1, &texID)
	if texID == 0 {
		log.Fatalln("[Video]: Failed to create the vid texture")
	}

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texID)
	coreRatioViewport(fbw, fbh)

	gl.BindVertexArray(0)

	e := gl.GetError()
	for e != gl.NO_ERROR {
		log.Printf("[Video] OpenGL error: %d\n", e)
		e = gl.GetError()
	}

}
