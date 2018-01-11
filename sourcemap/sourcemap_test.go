package sourcemap

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestCheckIfTrusted tests trustedURLs matches default SourceMapWhitelist pattern
func TestCheckIfTrusted(t *testing.T) {

	whiteListedURLs := []string{
		"http://localhost/test.js",
		"http://localhost/level_1/test.js",
		"http://localhost/level_2/level1/test.js",
		"http://localhost/level1_1/test1.js",
		"http://localhost/level_2/level1/test1.js",
		"https://localhost/test.js",
		"https://localhost/level_1/test.js",
		"https://localhost/level_2/level1/test.js",
		"https://localhost/level1_1/test1.js",
		"https://localhost/level_2/level1/test1.js",
	}

	blackListedURLs := []string{
		"http://localhost.com",
		"http://localhost@baleevemeiarelocalhost.com",
		"http://baleevemeiarelocalhost.com/myrequest#localhost\\",
		"https://localhost.com",
		"https://localhost@baleevemeiarelocalhost.com",
		"https://baleevemeiarelocalhost.com/myrequest#localhost\\",
	}

	testprocessor := Processor{
		Trusted: "^(http|https)://localhost/[^.]*",
	}

	testprocessor.Start()

	Convey("Using whiteListedURLs", t, func() {
		for _, URL := range whiteListedURLs {
			err := testprocessor.checkIfTrusted(URL)
			So(err, ShouldBeNil)
		}
	})

	Convey("Using blackListedURLs", t, func() {
		for _, URL := range blackListedURLs {
			err := testprocessor.checkIfTrusted(URL)
			So(err, ShouldBeError)
		}
	})

}
