package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
)

var errNoAction = errors.New("no action specified")

const sysDir = "/sys/class/backlight"

type cmdArgs struct {
	dec float64
	inc float64
	set float64
	get bool
	max bool
	min bool
}

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

func getValueFile(video string) string {
	return path.Join(video, "brightness")
}

func getCurrentValue(video string) (float64, error) {
	valueFile := getValueFile(video)
	currentValue, err := getNumberFromFile(valueFile)
	if err != nil {
		return 0, err
	}

	return currentValue, nil
}

func getMaxValue(video string) (float64, error) {
	maxValueFile := path.Join(video, "max_brightness")
	maxValue, err := getNumberFromFile(maxValueFile)
	if err != nil {
		return 0, err
	}

	return maxValue, nil
}

func calcNewValue(video string, args cmdArgs, maxValue float64) (float64, error) {
	const minValue float64 = 1000
	var newValue float64
	if args.set != 0 {
		newValue = getChangeValue(maxValue, args.set)
	} else if args.max {
		newValue = maxValue
	} else if args.min {
		newValue = minValue
	} else if args.inc != 0 {
		currentValue, err := getCurrentValue(video)
		if err != nil {
			return 0, err
		}
		newValue = currentValue + getChangeValue(maxValue, args.inc)
	} else if args.dec != 0 {
		currentValue, err := getCurrentValue(video)
		if err != nil {
			return 0, err
		}
		newValue = currentValue - getChangeValue(maxValue, args.dec)
	} else {
		return 0, errNoAction
	}

	return math.Max(minValue, math.Min(maxValue, math.Round(newValue))), nil
}

func setNewValue(video string, newValue float64) error {
	var mode os.FileMode
	return os.WriteFile(getValueFile(video), []byte(strconv.Itoa(int(newValue))), mode)
}

func handleCommand(video string, args cmdArgs) (string, error) {
	const minValue float64 = 1000
	maxValue, err := getMaxValue(video)
	if err != nil {
		return "", err
	}

	if args.get {
		currentValue, err := getCurrentValue(video)
		if err != nil {
			return "", err
		}
		if currentValue == 0 {
			fmt.Println(0)
			return "", nil
		}
		return fmt.Sprintf("%s: %1.2f%%", video, currentValue/maxValue*100), nil
	} else {
		newValue, err := calcNewValue(video, args, maxValue)
		if err != nil {
			if errors.Is(err, errNoAction) {
				return "", nil
			}
			return "", err
		}

		newValue = math.Max(minValue, math.Min(maxValue, math.Round(newValue)))
		setNewValue(video, newValue)

		return fmt.Sprintf("set %s to %d", video, int(newValue)), nil
	}

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
			output, err := handleCommand(file, args)
			if err != nil {
				fmt.Fprint(os.Stderr, err)
				os.Exit(1)
			}

			if output != "" {
				fmt.Println(output)
			}
		}
	}
}
