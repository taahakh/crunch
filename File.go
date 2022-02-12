package speed

import (
	"io"
	"log"
	"os"
)

func BufferByteRead(n string) []byte {

	f, err := os.Open(n)
	if err != nil {
		log.Fatal("AYOOO")
	}
	defer f.Close()

	stream := make([]byte, 0)
	b := make([]byte, 1024)

	for {
		_, err := f.Read(b)

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal("ok wtf")
		}
		stream = append(stream, b...)
	}
	return stream
}
