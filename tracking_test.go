package ups

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestRequests(t *testing.T) {

	tests := []struct {
		u, pwd, ak       string
		fixtureFile      string
		expectedErr      bool
		expectedActivity int
	}{
		{
			u:                "user",
			pwd:              "pwd",
			ak:               "access_key",
			fixtureFile:      "good-response.json",
			expectedActivity: 5,
		},
		{
			fixtureFile: "invalid-access-number.json",
		},
		{
			fixtureFile: "invalid-authentication-info.json",
		},
		{
			fixtureFile: "no-information-found.json",
		},
	}

	for _, tst := range tests {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if f, err := os.Open("fixtures/" + tst.fixtureFile); err == nil {
				w.WriteHeader(200)
				b, err := ioutil.ReadAll(f)
				if err != nil {
					panic(err)
				}
				w.Write(b)
				return
			}
		}))

		c := NewUPSTrackingClient(tst.u, tst.pwd, tst.pwd, nil)
		c.apiUrl = ts.URL

		res, err := c.TrackActivity("123")
		if err == nil && tst.expectedErr {
			t.Errorf("expected an err")
		}

		if err != nil {
			continue
		}
		if len(res.Shipment.Package.Activity) != tst.expectedActivity {
			t.Errorf("expected len(Activity): %d   got: %d", tst.expectedActivity, len(res.Shipment.Package.Activity))
		}

		if res.Shipment.ShipperNumber == "" {
			t.Errorf("expected a valid res.Shipment.ShipperNumber")
		}

		if len(res.Shipment.ShipmentAddress) == 0 {
			t.Errorf("expected a valid res.Shipment.ShipmentAddress")
			return
		}

		if res.Shipment.ShipmentAddress[0].Address.City == "" {
			t.Errorf("expected a valid res.Shipment.ShipmentAddress[0].Address.City")
		}

		if res.Shipment.ShipmentWeight.Weight == "" {
			t.Errorf("expected a valid res.Shipment.ShipmentWeight.Weight")
		}

		if len(res.Shipment.Package.PackageServiceOption) == 0 {
			t.Errorf("expected a valid len(res.Shipment.Package.PackageServiceOption)")
			return
		}
		if res.Shipment.Package.PackageServiceOption[0].Type.Code != "02" ||
			res.Shipment.Package.PackageServiceOption[0].Type.Description != "Adult Signature Required" {
			t.Errorf("expected a valid res.Shipment.Package.PackageServiceOption[0].Type: %#v", res.Shipment.Package.PackageServiceOption[0].Type)
		}

		if res.Shipment.Package.PackageWeight.Weight == "" {
			t.Errorf("expected a valid res.Shipment.Package.PackageWeight.Weight")
			return
		}

		if res.Shipment.Package.Message.Description == "" {
			t.Errorf("expected a valid res.Shipment.Package.Message.Description")
			return
		}

		ts.Close()
	}
}

func TestParsetime(t *testing.T) {
	a := []Activity{
		Activity{
			Date: "20180203",
			Time: "131300",
		},
	}

	parseDatetime(&a)

	want := time.Date(2018, 2, 3, 13, 13, 0, 0, time.UTC)
	if a[0].ParsedTimestamp != want {
		t.Errorf("wrong timestamp, got: %s  want: %s", a[0].ParsedTimestamp, want)
	}
}
