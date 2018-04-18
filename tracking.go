package ups

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Address struct {
	City              string
	StateProvinceCode string
	CountryCode       string
}

type CodeDescr struct {
	Code        string
	Description string
}

type Activity struct {
	ActivityLocation struct {
		Address Address
	}
	Status struct {
		CodeDescr
		Type string
	}
	Date            string
	Time            string
	ParsedTimestamp time.Time
}

type Weight struct {
	UnitOfMeasurement struct {
		Code string
	}
	Weight string
}

type TrackResponse struct {
	Response struct {
		ResponseStatus CodeDescr
	}
	Shipment struct {
		Package struct {
			Activity             []Activity
			PackageServiceOption []struct {
				Type CodeDescr
			}
			Message       CodeDescr
			PackageWeight Weight
		}
		ShipperNumber   string
		ShipmentAddress []struct {
			Type    CodeDescr
			Address Address
		}
		ShipmentWeight Weight
		Service        CodeDescr

		PickupDate string

		DeliveryDetail struct {
			Type CodeDescr
			Date string
		}
	}
}

type Fault struct {
	Faultcode   string
	Faultstring string
	Detail      struct {
		Errors struct {
			ErrorDetail struct {
				Severity         string
				PrimaryErrorCode CodeDescr
			}
		}
	} `json:"detail"`
}

type ResponseData struct {
	TrackResponse TrackResponse
	Fault         Fault
}

type RequestData struct {
	Security struct {
		UsernameToken struct {
			Username string
			Password string
		}
		UPSServiceAccessToken struct {
			AccessLicenseNumber string
		}
	}
	TrackRequest struct {
		Request struct {
			RequestAction string
			RequestOption string
		}
		InquiryNumber string
	}
}

type Client struct {
	Username   string
	Password   string
	AccessKey  string
	HTTPClient *http.Client
	apiUrl     string
}

func NewUPSTrackingClient(username, password, accessKey string, hc *http.Client) *Client {
	if hc == nil {
		hc = &http.Client{}
	}
	return &Client{
		Username:   username,
		Password:   password,
		AccessKey:  accessKey,
		HTTPClient: hc,
		apiUrl:     "https://onlinetools.ups.com/rest/Track",
	}
}

func parseDatetime(activities *[]Activity) {
	for idx, a := range *activities {
		s := fmt.Sprintf("%s-%s-%sT%s:%s:%sZ",
			a.Date[0:4], a.Date[4:6], a.Date[6:8],
			a.Time[0:2], a.Time[2:4], a.Time[4:6],
		)
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			(*activities)[idx].ParsedTimestamp = t
		}
	}
}

func (c *Client) TrackActivity(trackingNum string) (*TrackResponse, error) {
	reqData := RequestData{}
	reqData.Security.UsernameToken.Username = c.Username
	reqData.Security.UsernameToken.Password = c.Password
	reqData.Security.UPSServiceAccessToken.AccessLicenseNumber = c.AccessKey
	reqData.TrackRequest.Request.RequestAction = "Track"
	reqData.TrackRequest.Request.RequestOption = "activity"
	reqData.TrackRequest.InquiryNumber = trackingNum

	jsonStr, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.apiUrl, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res ResponseData
	if err = json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	if res.Fault.Faultcode != "" {
		return nil, fmt.Errorf(
			"UPS error, fault code: %s severity: %s  descr: %s",
			res.Fault.Faultcode,
			res.Fault.Detail.Errors.ErrorDetail.Severity,
			res.Fault.Detail.Errors.ErrorDetail.PrimaryErrorCode.Description)
	}

	// todo: confirm this is correct
	if res.TrackResponse.Response.ResponseStatus.Code != "1" {
		return nil, fmt.Errorf(
			"error, status code: %s descr: %s",
			res.TrackResponse.Response.ResponseStatus.Code,
			res.TrackResponse.Response.ResponseStatus.Description)
	}

	parseDatetime(&res.TrackResponse.Shipment.Package.Activity)

	return &res.TrackResponse, nil
}
