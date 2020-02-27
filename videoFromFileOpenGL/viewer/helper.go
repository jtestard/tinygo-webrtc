package main

import (
	"log"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/libretro/ludo/state"
)

// XYWHTo4points converts coordinates from (x, y, width, height) to (x1, y1, x2, y2, x3, y3, x4, y4)
func XYWHTo4points(x, y, w, h, fbh float32) (x1, y1, x2, y2, x3, y3, x4, y4 float32) {
	x1 = x
	x2 = x
	x3 = x + w
	x4 = x + w
	y1 = fbh - (y + h)
	y2 = fbh - y
	y3 = fbh - (y + h)
	y4 = fbh - y
	return
}

// coreRatioViewport configures the vertex array to display the game at the center of the window
// while preserving the original ascpect ratio of the game or core
func coreRatioViewport(fbWidth int, fbHeight int) (x, y, w, h float32) {
	// Scale the content to fit in the viewport.
	fbw := float32(fbWidth)
	fbh := float32(fbHeight)

	// NXEngine workaround
	aspectRatio := float32(Geom.AspectRatio)
	if aspectRatio == 0 {
		aspectRatio = float32(Geom.BaseWidth) / float32(Geom.BaseHeight)
	}

	h = fbh
	w = fbh * aspectRatio
	if w > fbw {
		h = fbw / aspectRatio
		w = fbw
	}

	// Place the content in the middle of the window.
	x = (fbw - w) / 2
	y = (fbh - h) / 2

	va := vertexArray(x, y, w, h, 1.0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(va)*4, gl.Ptr(va), gl.STATIC_DRAW)

	return
}

func vertexArray(x, y, w, h, scale float32) []float32 {
	fbw, fbh := window.GetFramebufferSize()
	ffbw := float32(fbw)
	ffbh := float32(fbh)

	w *= scale
	h *= scale

	x1, y1, x2, y2, x3, y3, x4, y4 := XYWHTo4points(x, y, w, h, ffbh)

	return []float32{
		//  X, Y, U, V
		x1/ffbw*2 - 1, y1/ffbh*2 - 1, 0, 1, // left-bottom
		x2/ffbw*2 - 1, y2/ffbh*2 - 1, 0, 0, // left-top
		x3/ffbw*2 - 1, y3/ffbh*2 - 1, 1, 1, // right-bottom
		x4/ffbw*2 - 1, y4/ffbh*2 - 1, 1, 0, // right-top
	}
}

// InitFramebuffer initializes and configures the video frame buffer based on
// informations from the HWRenderCallback of the libretro core.
func initFramebuffer(width, height int) {
	log.Printf("[Video]: Initializing HW render (%v x %v).\n", width, height)

	gl.GenFramebuffers(1, &fboID)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fboID)

	//gl.GenTextures(1, &video.texID)
	gl.BindTexture(gl.TEXTURE_2D, texID)
	gl.TexStorage2D(gl.TEXTURE_2D, 1, gl.RGBA8, int32(width), int32(height))

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texID, 0)

	hw := state.Global.Core.HWRenderCallback

	gl.BindRenderbuffer(gl.RENDERBUFFER, 0)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		log.Fatalln("[Video] Framebuffer is not complete.")
	}

	gl.ClearColor(0, 0, 0, 1)
	if hw.Depth && hw.Stencil {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	} else if hw.Depth {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	} else {
		gl.Clear(gl.COLOR_BUFFER_BIT)
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}
