package utils

import (
	"log"
)

// LogError writes the error to the log
func LogError(err error) {
	if err != nil {
		log.Print(err)
	}
}

// LogErrorF writes the error to the log. Fails if there is an error
func LogErrorF(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
