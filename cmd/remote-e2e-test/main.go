package main

// type testingT struct{}

// func (testingT) Errorf(format string, args ...interface{}) {
// 	fmt.Printf(format, args...)
// }

// func (testingT) FailNow() {
// 	// https://www.youtube.com/watch?v=RlnlDKznIaw <-- It's NECK n' NECK!
// 	panic("B00M!! TETRIS FOR JEFF!")
// }

// func main() {
// 	c := config.New()

// 	defaultAPIURL := c.Host
// 	if defaultAPIURL[0:1] == `:` {
// 		defaultAPIURL = "http://localhost" + defaultAPIURL
// 	}

// 	apiURL := pflag.String("url", defaultAPIURL, "CodeCoach API URL, default: http://localhost:9001")

// 	pflag.Parse()

// 	t := new(testing.T)
// 	ts := new(e2e.TS)
// 	ts.APIURL = apiURL

// 	ts.TestAll(t)

// 	fmt.Printf("ALL TESTS PASS!\n")
// }
