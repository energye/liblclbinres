//----------------------------------------
//
// Copyright © yanghy. All Rights Reserved.
//
// Licensed under Apache License Version 2.0, January 2004
//
// https://www.apache.org/licenses/LICENSE-2.0
//
//----------------------------------------

package main

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"fmt"
	"github.com/energye/liblclbinres/v2/genbinres/home"
	"hash/crc32"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var liblclVersion = "v0.0.0" // liblcl发布版本

const (
	golcl = "golcl"
)

func main() {
	if len(os.Args) > 1 {
		liblclVersion = os.Args[1]
		if liblclVersion[0] != 'v' {
			println("liblcl version is error")
			os.Exit(1)
		}
	} else {
		println("liblcl version is nil")
		os.Exit(1)
	}
	wd, _ := os.Getwd()
	libLCLBinResDir := filepath.Join(wd, "../")
	if strings.LastIndex(libLCLBinResDir, "liblclbinres") == -1 {
		libLCLBinResDir = filepath.Join(libLCLBinResDir, "liblclbinres")
	}
	println("liblcl version:", liblclVersion)
	println("out liblcl dir:", libLCLBinResDir)
	// 用户目录
	dir, err := home.Dir()
	if err != nil {
		panic(err)
	}
	liblclPath := filepath.Join(dir, golcl)
	fmt.Println("用户目录:", dir)
	finfo, err := ioutil.ReadDir(liblclPath)
	if err != nil {
		panic(err)
	}

	for _, info := range finfo {
		zipPath := filepath.Join(liblclPath, info.Name())
		zz, err := zip.OpenReader(zipPath)
		if err != nil {
			println("open-zip-error:", err.Error())
			continue
		}
		defer zz.Close()
		var (
			file fs.File
		)
		name := strings.ToLower(info.Name())
		if strings.Contains(name, "windows") {
			file, err = zz.Open("liblcl.dll")
		} else if strings.Contains(name, "linux") {
			file, err = zz.Open("liblcl.so")
		} else if strings.Contains(name, "macos") {
			file, err = zz.Open("liblcl.dylib")
		}
		if err != nil {
			println("open-dll-error:", err.Error())
			continue
		}
		defer file.Close()
		data, err := ioutil.ReadAll(file)
		if err != nil {
			println("read-dll-error:", err.Error())
			continue
		}
		if strings.Contains(name, "liblcl.windows32.zip") {
			// windows 32
			genresByte(data, "windows && latest", filepath.Join(libLCLBinResDir, "liblcl_windows_386.go"))
		} else if strings.Contains(name, "liblcl.windows64.zip") {
			// windows 64
			genresByte(data, "windows && latest", filepath.Join(libLCLBinResDir, "liblcl_windows_amd64.go"))
		} else if strings.Contains(name, "liblcl.windowsarm64.zip") {
			// windows arm64
			genresByte(data, "windows && latest", filepath.Join(libLCLBinResDir, "liblcl_windows_arm64.go"))
		} else if strings.Contains(name, "liblcl-109.windows32.zip") {
			// windows - 109 32
			genresByte(data, "windows && 109", filepath.Join(libLCLBinResDir, "liblcl_windows7_386.go"))
		} else if strings.Contains(name, "liblcl-109.windows64.zip") {
			// windows - 109 64
			genresByte(data, "windows && 109", filepath.Join(libLCLBinResDir, "liblcl_windows7_amd64.go"))
		} else if strings.Contains(name, "liblcl.linux64.zip") {
			// linux 64
			genresByte(data, "linux && latest", filepath.Join(libLCLBinResDir, "liblcl_gtk3_linux_amd64.go"))
		} else if strings.Contains(name, "liblcl.linuxarm64.zip") {
			// linux arm64
			genresByte(data, "linux && latest", filepath.Join(libLCLBinResDir, "liblcl_gtk3_linux_arm64.go"))
		} else if strings.Contains(name, "liblcl.linux64gtk2.zip") {
			// linux - 106 64
			genresByte(data, "linux && 106", filepath.Join(libLCLBinResDir, "liblcl_gtk2_linux_amd64.go"))
		} else if strings.Contains(name, "liblcl.linuxarm64gtk2.zip") {
			// linux - 106 arm64
			genresByte(data, "linux && 106", filepath.Join(libLCLBinResDir, "liblcl_gtk2_linux_arm64.go"))
		} else if strings.Contains(name, "liblcl.macosarm64.zip") {
			// macos arm64
			genresByte(data, "darwin && latest", filepath.Join(libLCLBinResDir, "liblcl_darwin_arm64.go"))
		} else if strings.Contains(name, "liblcl.macosx64.zip") {
			// macosx 64
			genresByte(data, "darwin && latest", filepath.Join(libLCLBinResDir, "liblcl_darwin_amd64.go"))
		}
	}
	genresLiblclVersion(libLCLBinResDir, liblclVersion)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// zlib压缩
func zlibCompress(input []byte) ([]byte, error) {
	var in bytes.Buffer
	w, err := zlib.NewWriterLevel(&in, zlib.BestCompression)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(input)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return in.Bytes(), nil
}

func genresByte(input []byte, tags, newFileName string) {
	fmt.Println("genFile: ", newFileName)
	if len(input) == 0 {
		fmt.Println("000000")
		return
	}
	crc32Val := crc32.ChecksumIEEE(input)
	//压缩
	bs, err := zlibCompress(input)
	if err != nil {
		panic(err)
	}
	code := bytes.NewBuffer(nil)
	code.WriteString("//go:build ")
	code.WriteString(tags)
	code.WriteString("\r\n\r\n")
	code.WriteString("package liblclbinres")
	code.WriteString("\r\n\r\n")
	code.WriteString(fmt.Sprintf("const CRC32Value uint32 = 0x%x\r\n\r\n", crc32Val))

	code.WriteString("var LCLBinRes = []byte(\"")
	for _, b := range bs {
		code.WriteString("\\x" + fmt.Sprintf("%.2x", b))
	}
	code.WriteString("\")\r\n")
	ioutil.WriteFile(newFileName, code.Bytes(), 0666)
}

func genresLiblclVersion(libLCLBinResDir, version string) {
	code := bytes.NewBuffer(nil)
	code.WriteString("package liblclbinres")
	code.WriteString("\r\n\r\n")
	code.WriteString(`const version = "` + version + `"`)
	code.WriteString("\r\n\r\n")
	code.WriteString("func LibVersion() string {")
	code.WriteString("\n\t")
	code.WriteString("return version")
	code.WriteString("\n}")
	ioutil.WriteFile(filepath.Join(libLCLBinResDir, "liblcl.go"), code.Bytes(), 0666)
}
