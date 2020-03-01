package main

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
)

func shaderLog(shader uint32) string {
	var length, chWritten int32
	gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &length)
	var msg []uint8
	if length > 0 {
		msg = make([]uint8, length)
		gl.GetShaderInfoLog(shader, length, &chWritten, &msg[0])
		fmt.Printf("shader log: %s\n", goString(msg))
	}
	return goString(msg)
}

func programLog(program uint32) string {
	var length, chWritten int32
	gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &length)
	var msg []uint8
	if length > 0 {
		msg = make([]uint8, length)
		gl.GetProgramInfoLog(program, length, &chWritten, &msg[0])
		fmt.Printf("program log: %s\n", goString(msg))
	}
	return goString(msg)
}

func goString(buf []uint8) string {
	for i, b := range buf {
		if b == 0 {
			return string(buf[:i])
		}
	}
	panic("buf is not NUL-terminated")
}

func checkOpenGLError() {
	foundErr := false
	glErr := gl.GetError()
	for glErr != gl.NO_ERROR {
		fmt.Printf("OpenGL Error Code: %X\n", glErr)
		foundErr = true
		glErr = gl.GetError()
	}
	if foundErr {
		// panic("open gl error found, closing...")
	}
}
