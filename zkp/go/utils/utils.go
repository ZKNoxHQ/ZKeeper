package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// writeToFile is a helper to serialize and write gnark objects or byte readers to files.
func WriteToFile(filename string, data interface{}) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", filename, err)
		os.Exit(1)
	}
	defer file.Close()

	switch v := data.(type) {
	case io.WriterTo:
		_, err = v.WriteTo(file)
	case *bytes.Reader: // For the JSON input
		_, err = v.WriteTo(file)
	default:
		err = fmt.Errorf("unsupported type for writing to file")
	}

	if err != nil {
		fmt.Printf("Error writing to file %s: %v\n", filename, err)
		os.Exit(1)
	}
	fmt.Printf("Wrote %s\n", filename)
}

// readFromFile is a helper to deserialize and read gnark objects or JSON from files.
func ReadFromFile(filename string, data interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", filename, err)
	}
	defer file.Close()

	switch v := data.(type) {
	case io.ReaderFrom:
		_, err = v.ReadFrom(file)
		if err != nil && err != io.EOF { // io.EOF is expected if the file is empty or partially read
			return fmt.Errorf("error reading from file %s into io.ReaderFrom: %w", filename, err)
		}
	default:
		return fmt.Errorf("unsupported type for reading from file: %T", data)
	}

	return nil
}
