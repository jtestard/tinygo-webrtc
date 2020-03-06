package main

import (
	"fmt"
	"github.com/go-gl/glfw/v3.3/glfw"
	"path/filepath"
)

type Drawer interface {
	LoadTexture(prog uint32) error
	LoadProgram(prog uint32) error
	DrawScene(vao uint32, window *glfw.Window, program uint32)
}

func NewDrawer(file string) (Drawer, error) {
	ext := filepath.Ext(file)
	switch ext {
	case ".png":
		fallthrough
	case ".jpeg":
		return &ImgDrawer{file: file}, nil
	case ".ivf":
		return &VideoDrawer{file: file}, nil
	default:
		return nil, fmt.Errorf("cannot load file with extension: %s", ext)
	}
}
