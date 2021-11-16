package grib2

import (
	"fmt"
	"os"
	"path"
	"testing"

	"h12.io/wipro"
)

func TestRead(t *testing.T) {
	const testDir = "/Users/w/dq/voyage-monitor/weather-stream-service/tools/spire-cli/"
	const filename = "sof-d.20211115.t00z.0p125.basic.global.f000.grib2"
	f, err := os.Open(path.Join(testDir, filename))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	r := wipro.NewFileReader(f)

	var msg Message

	msg.Unmarshal(r)
	if r.Err() != nil {
		t.Fatal(r.Err())
	}

	fmt.Printf("%+v\n", &msg)
}
