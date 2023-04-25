package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

func getIntValFromFile(fileName string) (int, error) {
	f, err := os.ReadFile(fileName)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(f)))
}

func getChangeValue(maxValue, change int) int {
	return (change * maxValue / 100)
}

const sysDir = "/sys/class/backlight"

func getVideoPath() (string, error) {
	video := os.Getenv("VIDEO_DEVPATH")
	if video != "" {
		return video, nil
	}
	files, err := os.ReadDir(sysDir)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", errors.New(fmt.Sprintf("no files found in %s", sysDir))
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	return path.Join(sysDir, files[0].Name()), nil
}

func main() {
	inc := flag.Int("inc", 0, "percentage")
	dec := flag.Int("dec", 0, "percentage")
	max := flag.Bool("max", false, "set max brightness")
	min := flag.Bool("min", false, "set min brightness")
	get := flag.Bool("get", false, "get current percentage")
	flag.Parse()

	video, err := getVideoPath()

	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	valueFile := path.Join(video, "brightness")
	maxFile := path.Join(video, "max_brightness")
	minValue := 1000
	currentValue, err := getIntValFromFile(valueFile)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	maxValue, err := getIntValFromFile(maxFile)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	var newValue int
	if *get {
		if currentValue == 0 {
			fmt.Println(0)
			return
		}
		fmt.Println(maxValue / currentValue * 100)
		return
	} else if *max {
		newValue = maxValue
	} else if *min {
		newValue = minValue
	} else if *inc != 0 {
		newValue = currentValue + getChangeValue(maxValue, *inc)
	} else if *dec != 0 {
		newValue = currentValue - getChangeValue(maxValue, *dec)
	} else {
		return
	}

	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	var mode os.FileMode
	if err := os.WriteFile(valueFile, []byte(strconv.Itoa(newValue)), mode); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
