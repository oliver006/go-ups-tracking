package main

import(
	"fmt"
	"flag"

	ups "github.com/oliver006/go-ups-tracking"
	"os"
)

func printUsage() {
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("")
	fmt.Println("cli << tracking-number>>")
	fmt.Println("")
	fmt.Println("Environment variables:")
	fmt.Println("UPS_USERNAME, UPS_PASSWORD, UPS_ACCESS_KEY")
}

func main() {

	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Println("not enough parameters")
		printUsage()
		return
	}

	tn := flag.Arg(0)

	c := ups.NewUPSTrackingClient(
		os.Getenv("UPS_USERNAME"),
		os.Getenv("UPS_PASSWORD"),
		os.Getenv("UPS_ACCESS_KEY"),
		nil,
	)

	res, err := c.TrackActivity(tn)
	if err != nil {
		fmt.Printf("error: %s \n", err)
		return
	}

	fmt.Printf("Tracking Number: %s \n", res.Shipment.Package.TrackingNumber)

	fmt.Println()
	fmt.Println("Package Activity")
	fmt.Println("================")
	fmt.Println()
	for _, a := range res.Shipment.Package.Activity {
		fmt.Println(a.String())
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("Package Info")
	fmt.Println("============")
	fmt.Println()

	fmt.Printf("Service: %s\n", res.Shipment.Service.String())
	fmt.Printf("Shipper: %s\n", res.Shipment.ShipperNumber)
	fmt.Printf("Weight: %s %s", res.Shipment.ShipmentWeight.Weight, res.Shipment.ShipmentWeight.UnitOfMeasurement.Code)
	fmt.Println()
	fmt.Println()
}

