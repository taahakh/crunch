package crunch

import (
	"encoding/csv"
	"errors"
	"log"
	"os"
)

func ReadCSV(file string) (csvReader [][]string, err error) {
	// we are opening the csv file, checking if it exists and making sure it closes in the end
	csvFile, err := os.Open(file)
	if err != nil {
		log.Println(err)
	}
	defer csvFile.Close()

	// reads the csv files in rows. each row is in an array
	csvReader, err = csv.NewReader(csvFile).ReadAll()
	if err != nil {
		log.Println(err)
	}

	return csvReader, err
}

func WriteCSV(file string, records [][]string) (state bool, err error) {

	// creates the file
	csvFile, err := os.Create(file)
	if err != nil {
		log.Println("Couldn't open")
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	// writes each record in the file
	// state tells the end user if this transactions was sucessful
	state = true
	for _, record := range records {
		if csvWriter.Write(record); err != nil {
			state = false
			log.Println(err)
		}
	}

	return state, err
}

// checks existsece, returns state and any errors thrown
// checks for directory/file
func Exists(file string) (bool, error) {
	_, err := os.Stat(file)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
