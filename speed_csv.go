package speed

import (
	"encoding/csv"
	"io"
	"log"
	"os"
)

func ReadPCCSV(file string) (csvReader [][]string) {

	csvFile, err := os.Open(file)
	if err != nil {
		log.Println(err, "---ReadPCCSV-os.Open")
	}
	defer csvFile.Close()

	r := csv.NewReader(csvFile)
	// csvReader, err = r.ReadAll()
	// if err != nil {
	// 	log.Println("ok, breakage")
	// }
	// csvReader = make([][]string, 0)

	for {
		record, err := r.Read()
		// breaks out of while loop when reaches end of file
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("This shit aint working")
		}
		// item := [][]string{{record[0]}}
		csvReader = append(csvReader, []string{record[1]})

	}

	return csvReader

}

func WritePCCSV(file string, cvList [][]string) {
	// creates the file
	csvFile, err := os.Create(file)
	if err != nil {
		log.Println("Couldn't open")
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	err = csvWriter.WriteAll(cvList)
}

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

func WriteCSVpointer(file string, rec **[][]string) {
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
	// state = true
	for _, record := range **rec {
		if csvWriter.Write(record); err != nil {
			log.Println(err)
		}
	}

}
