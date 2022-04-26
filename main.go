package main

import (
	"fmt"
	"image"
	"strings"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"github.com/disintegration/imaging"
	"gopkg.in/yaml.v2"
	"time"
	"sync"
)

const (
	ConfFile = "conf.yaml"
	OutDir = "out"
	InDir = "in"
)

var anchorMap = map[string]imaging.Anchor{
	"center": imaging.Center,
	"topleft": imaging.TopLeft,
	"top": imaging.Top,
	"topright": imaging.TopRight,
	"left": imaging.Left,
	"right": imaging.Right,
	"bottomleft": imaging.BottomLeft,
	"bottom": imaging.Bottom,
	"bottomRight": imaging.BottomRight,
}

var wg sync.WaitGroup

type Conf struct {
	Sizes []Size `yaml:"sizes"`
}

type Size struct {
	Name string `yaml:"name"`
	Width int `yaml:"width"`
	Height int `yaml:"height"`
	Quality int `yaml:"quality"`
	Mode string `yaml:"mode"`
	Anchor string `yaml:"anchor"`
}

func getConf() (*Conf, error) {
    yamlFile, err := ioutil.ReadFile(ConfFile)
    if err != nil {
		log.Printf("Error reading file %v", ConfFile)
        return nil, err
    }
	c := &Conf{}
    err = yaml.Unmarshal(yamlFile, c)
    if err != nil {
		log.Printf("Error unmarshalling file %v", ConfFile)
        return nil, fmt.Errorf("Unmarshal: %v", err)
    }
	return c, nil
}

func makeDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, os.ModeDir|0755)
	}
	return nil
}

func checkDirectoryIfExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func isImage(ext string) bool {
	return (strings.ToLower(ext) == ".jpg" ||
	strings.ToLower(ext) == ".jpeg" ||
	strings.ToLower(ext) == ".png")
}

func readFileFromDir() []string {
	var files []string

    err := filepath.Walk(InDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && isImage(filepath.Ext(path)) {
        	files = append(files, info.Name())
		}
        return nil
    })
    if err != nil {
        panic(err)
    }
	return files
}

func timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    log.Printf("%s took %s", name, elapsed)
}

func processImage(size Size, file string) {
	defer wg.Done()

	src, err := imaging.Open(InDir+"/"+file)
	if err != nil {
		log.Fatalf("Failed to open image: %v", err)
	}

	var dst *image.NRGBA
	switch size.Mode {
	case "crop":
		dst = imaging.CropAnchor(src, size.Width, size.Height, anchorMap[size.Anchor])
	case "fill":
		dst = imaging.Fill(src, size.Width, size.Height, anchorMap[size.Anchor], imaging.Lanczos)
	case "fit":
		dst = imaging.Fit(src, size.Width, size.Height, imaging.Lanczos)
	}
	
	err = imaging.Save(dst, OutDir+"/"+size.Name+"_"+file, imaging.JPEGQuality(size.Quality))
	if err != nil {
		log.Fatalf("Failed to save image: %v", err)
	}
}

func main() {
	log.Println("Check if input dir exists...")
	if !checkDirectoryIfExists(InDir) {
		makeDirectoryIfNotExists(InDir)
		log.Fatalln("Input directory does not exist, making one for you")
	}

	log.Println("Reading configuration file...")
	
	conf, err := getConf()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Making output dir...")
	makeDirectoryIfNotExists(OutDir)


	log.Printf("Reading file inside %v dir...", InDir)
	files := readFileFromDir()
	log.Printf("Found %v files", len(files))

	defer timeTrack(time.Now(), "processing")
	for i, size := range conf.Sizes {
		log.Printf("Generating size %v, size name: %v...", i, size.Name)

		for _, file := range files {
			log.Println("Working with file", InDir+"/"+file)
			wg.Add(1)
			go processImage(size, file)
		}
	}

	wg.Wait()
}