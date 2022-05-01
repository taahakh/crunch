package response

// func Run() {
// 	http.HandleFunc("/", handler)
// 	http.HandleFunc("/start", startscraping)
// 	http.ListenAndServe(":3000", nil)
// }

// func handler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("Getting resposen")
// 	w.WriteHeader(200)
// 	w.Write([]byte("ok"))
// }

// func startscraping(w http.ResponseWriter, r *http.Request) {

// 	csv, err := speed.ReadCSV("C:\\Users\\taaha\\go\\src\\github.com\\taahakh\\speed\\data\\req\\list.csv")
// 	if err != nil {
// 		log.Println("There is no file")
// 	}

// 	t, err := time.ParseDuration("2s")
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	rj1 := req.SimpleProxySetup(
// 		req.GenodeRead(csv, "http"),
// 		[]string{
// 			"https://httpbin.org",
// 			"https://ruktaj.co.uk",
// 		},
// 		nil,
// 		2,
// 		t,
// 		nil,
// 	)

// 	pool := req.Pool{}
// 	pool.SetName("main", nil)
// 	pool.Add("A", rj1)
// 	pool.Run("A", req.METHOD_COMPLETE, 2)
// }
