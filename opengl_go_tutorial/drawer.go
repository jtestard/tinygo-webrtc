package main

import (
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"path/filepath"
)

type Drawer interface {
	LoadTexture() error
	LoadProgram(prog uint32) error
	DrawScene(vao uint32, window *glfw.Window, program uint32)
}

func NewDrawer(file string) (Drawer, error) {
	ext := filepath.Ext(file)
	switch ext {
	case ".png":
		fallthrough
	case ".jpeg":
		imgDrawer := &ImgDrawer{file: file}
		return imgDrawer, nil
	case ".ivf":
		return nil, fmt.Errorf("cannot load file with extension: %s", ext)
	default:
		return nil, fmt.Errorf("cannot load file with extension: %s", ext)
	}
}
