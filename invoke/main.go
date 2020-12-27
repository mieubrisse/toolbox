package main

import "flag"

func main() {
	flag.String(
		"template",
		"",
		"Template to copy")
	flag.String(
		"destination",
		"",
		"Destination for filled template")
}
