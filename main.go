package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

const sysDir = "/sys/class/backlight"

func getNumberFromFile(fileName string) (float64, error) {
	f, err := os.ReadFile(fileName)
	if err != nil {
		return 0, err
	}
	intval, err := strconv.Atoi(strings.TrimSpace(string(f)))

	if err != nil {
		return 0, err
	}

	return float64(intval), nil
}

func getChangeValue(maxValue, change float64) float64 {
	return (change * maxValue / 100)
}

type cmdArgs struct {
	dec float64
	inc float64
	set float64
	get bool
	max bool
	min bool
}

func handleCommand(video string, args cmdArgs) {
	valueFile := path.Join(video, "brightness")
	maxFile := path.Join(video, "max_brightness")
	var minValue float64
	minValue = 1000
	currentValue, err := getNumberFromFile(valueFile)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	maxValue, err := getNumberFromFile(maxFile)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	var newValue float64
	if args.get {
		if currentValue == 0 {
			fmt.Println(0)
			return
		}
		fmt.Println(fmt.Sprintf("%s: %2.f%%", video, currentValue/maxValue*100))
		return
	} else if args.set != 0 {
		newValue = getChangeValue(maxValue, args.set)
	} else if args.max {
		newValue = maxValue
	} else if args.min {
		newValue = minValue
	} else if args.inc != 0 {
		newValue = currentValue + getChangeValue(maxValue, args.inc)
	} else if args.dec != 0 {
		newValue = currentValue - getChangeValue(maxValue, args.dec)
	} else {
		return
	}

	var mode os.FileMode
	if err := os.WriteFile(valueFile, []byte(strconv.Itoa(int(newValue))), mode); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(fmt.Sprintf("set %s to %d", video, int(newValue)))
}

func getVideoPaths() ([]string, error) {
	var out []string

	files, err := os.ReadDir(sysDir)
	if err != nil {
		return out, err
	}

	if len(files) == 0 {
		return out, errors.New(fmt.Sprintf("no files found in %s", sysDir))
	}

	for _, file := range files {
		out = append(out, path.Join(sysDir, file.Name()))
	}
	return out, nil

}

func main() {
	inc := flag.Float64("inc", 0, "percentage")
	dec := flag.Float64("dec", 0, "percentage")
	set := flag.Float64("set", 0, "percentage")
	max := flag.Bool("max", false, "set max brightness")
	min := flag.Bool("min", false, "set min brightness")
	get := flag.Bool("get", false, "get current percentage")
	dev := flag.String("dev", "", fmt.Sprintf("update only the device (listed in %s)", sysDir))
	list := flag.Bool("list", false, "list devices")
	flag.Parse()

	files, err := getVideoPaths()

	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	if *list {
		for _, f := range files {
			fmt.Println(f)
		}
		return
	}

	args := cmdArgs{
		dec: *dec,
		inc: *inc,
		set: *set,
		get: *get,
		max: *max,
		min: *min,
	}

	for _, file := range files {
		if *dev == "" || strings.HasSuffix(file, "/"+*dev) {
			handleCommand(file, args)
		}
	}
}
