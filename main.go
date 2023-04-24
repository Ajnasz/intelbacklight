package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const (
	valueFile = "/sys/class/backlight/intel_backlight/brightness"
	maxFile   = "/sys/class/backlight/intel_backlight/max_brightness"
)

func getIntValFromFile(fileName string) (int, error) {
	f, err := ioutil.ReadFile(fileName)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(f)))
}

func getIncValue(currentValue, change int) int {
	return currentValue + (change * currentValue / 100)
}

func getDecValue(currentValue, change int) int {
	return currentValue - (change * currentValue / 100)
}

func getNewValue(val string, currentValue, change, maxValue, minValue int) (int, error) {
	var newValue int
	switch val {
	case "inc":
		newValue = getIncValue(currentValue, change)
	case "dec":
		newValue = getDecValue(currentValue, change)
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

func main() {
	val := flag.String("val", "", "one of inc, dec, max, min")
	flag.Parse()

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

	change := 10

	newValue, err := getNewValue(*val, currentValue, change, maxValue, minValue)

	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	var mode os.FileMode
	if err := ioutil.WriteFile(valueFile, []byte(strconv.Itoa(newValue)), mode); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
