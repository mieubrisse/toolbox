package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const (
	itemDelimiter = ","
	rangeDelimiter = "-"
	ttyFilepath = "/dev/tty"
)

func main() {
	ttyOutFp, err := os.OpenFile(ttyFilepath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't open TTY for writing messages\n")
		os.Exit(1)
	}
	ttyInFp, err := os.OpenFile(ttyFilepath, os.O_RDONLY, 0644)
	if err != nil {
		ttyOutFp.WriteString("Couldn't open TTY for reading input\n")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	scanner := bufio.NewScanner(reader)
	lines := []string{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		ttyOutFp.WriteString("No results\n")
	}

	for idx, line := range lines {
		ttyOutFp.WriteString(fmt.Sprintf("%v\t%v\n", idx, line))
	}
	maxIndex := len(lines) - 1

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		os.Exit(1)
	}()

	var selectedIndices []int
	for selectedIndices == nil {
		ttyOutFp.WriteString("Use which?\n")
		var rawInput string
		_, err := fmt.Fscanln(ttyInFp, &rawInput)
		if err != nil {
			ttyOutFp.WriteString("An error occurred reading input string: " + err.Error())
			continue
		}
		selectedIndices, err = parseChoicesStr(rawInput, maxIndex)
		if err != nil {
			ttyOutFp.WriteString(err.Error() + "\n")
			continue
		}
	}

	for _, idx := range selectedIndices {
		fmt.Println(lines[idx])
	}
}


// Parse the raw choices string into a set of integers to select
func parseChoicesStr(str string, maxIdx int) ([]int, error) {
	selectedIndices := []int{}
	fragments := strings.Split(str, itemDelimiter)
	for _, fragment := range fragments {
		strippedFragment := strings.TrimSpace(fragment)

		// Ignore empty selections
		if len(strippedFragment) == 0 {
			continue
		}

		rangeElems := strings.Split(strippedFragment, rangeDelimiter)
		numRangeElems := len(rangeElems)
		if numRangeElems == 1 {
			elem := rangeElems[0]
			index, err := strconv.Atoi(elem)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Index '%v' not a number", elem))
			}
			if index < 0 || index > maxIdx {
				return nil, errors.New(fmt.Sprintf("Index '%v' must be >= 0 and <= %v", index, maxIdx))
			}
			selectedIndices = append(selectedIndices, index)
		} else if numRangeElems == 2 {
			rangeStart := rangeElems[0]
			rangeEnd := rangeElems[1]
			rangeStartInt, err := strconv.Atoi(rangeStart)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Range '%v' start isn't a number", strippedFragment))
			}
			rangeEndInt, err := strconv.Atoi(rangeEnd)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Range '%v' end isn't a number", strippedFragment))
			}
			if rangeStartInt < 0 {
				return nil, errors.New(fmt.Sprintf("Range '%v' start must be >= 0", strippedFragment))
			}
			if rangeStartInt >= rangeEndInt {
				return nil, errors.New(fmt.Sprintf("Range '%v' start cannot be >= end", strippedFragment))
			}
			if rangeEndInt > maxIdx {
				return nil, errors.New(fmt.Sprintf("Range '%v' end must be <= %v", strippedFragment, maxIdx))
			}
			for i := rangeStartInt; i <= rangeEndInt; i++ {
				selectedIndices = append(selectedIndices, i)
			}
		} else {
			return nil, errors.New(fmt.Sprintf("Invalid range '%v'", strippedFragment))
		}
	}
	if len(selectedIndices) == 0 {
		return nil, errors.New("At least one index must be selected")
	}
	return selectedIndices, nil
}
