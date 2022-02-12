package speed

// func BufferByteRead(f *os.File) []byte {

// 	defer f.Close()

// 	stream := make([]byte, 0)
// 	b := make([]byte, 1024)

// 	for {
// 		_, err := f.Read(b)

// 		if err == io.EOF {
// 			break
// 		}

// 		if err != nil {
// 			log.Fatal("ok wtf")
// 		}
// 		stream = append(stream, b...)
// 	}
// 	return stream
// }
