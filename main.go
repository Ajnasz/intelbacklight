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

func getNewValue(val string, currentValue, change, maxValue, minValue int) (int, error) {
	var newValue int
	switch val {
	case "inc":
		newValue = currentValue + getChangeValue(maxValue, change)
	case "dec":
		newValue = currentValue - getChangeValue(maxValue, change)
	case "max":
		newValue = maxValue
	case "min":
		newValue = minValue
	default:
		return 0, errors.New("invalid operation")
	}

	if newValue > maxValue {
		newValue = maxValue
	}

	if newValue < minValue {
		newValue = minValue
	}

	return newValue, nil
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

	sortableFiles := dirEntryByName(files)

	sort.Sort(sortableFiles)
	return path.Join(sysDir, sortableFiles[0].Name()), nil
}

func main() {
	val := flag.String("val", "", "one of inc, dec, max, min")
	change := flag.Int("change", 10, "percentage of change")
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

	newValue, err := getNewValue(*val, currentValue, *change, maxValue, minValue)

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
