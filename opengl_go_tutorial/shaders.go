package main

import (
	"fmt"
	_ "image/png"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
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

// textures are stored "upside-down" in open gl
func texInvertY(texCoords []float32) {
	if len(texCoords)%2 != 0 || len(texCoords) < 2 {
		return
	}
	for i := 1; i < len(texCoords); i += 2 {
		texCoords[i] = 1 - texCoords[i]
	}
}
