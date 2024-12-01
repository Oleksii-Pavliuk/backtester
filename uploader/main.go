package main

import (
	"uploader/processor"
)

func main() {
	processor.NewProcessor("files", "localhost", 3030)
}
