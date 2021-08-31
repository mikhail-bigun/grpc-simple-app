package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/mikhail-bigun/grpc-app-pcbook/client"
	"github.com/mikhail-bigun/grpc-app-pcbook/pb/pcbook"
	"github.com/mikhail-bigun/grpc-app-pcbook/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func testRateLaptop(laptopClient *client.LaptopClient) {
	n := 3
	laptopIDs := make([]string, n)

	for i := 0; i < n; i++ {
		laptop := sample.NewLaptop()
		laptopIDs[i] = laptop.Id
		laptopClient.CreateLaptop(laptop)
	}

	scores := make([]float64, n)
	for {
		fmt.Print("rate laptop (y/n)?")
		var answer string
		fmt.Scan(&answer)
		if strings.ToLower(answer) != "y" {
			break
		}

		for i := 0; i < n; i++ {
			scores[i] = sample.RandomLaptopScore()
		}

		err := laptopClient.RateLaptop(laptopIDs, scores)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func testCreateLaptop(laptopClient *client.LaptopClient) {
	laptopClient.CreateLaptop(sample.NewLaptop())
}

func testSearchLaptop(laptopClient *client.LaptopClient) {
	// Creating 10 random laptops
	for i := 0; i <= 10; i++ {
		testCreateLaptop(laptopClient)
	}

	// making filter
	filter := &pcbook.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam: &pcbook.Memory{
			Value: 8,
			Unit:  pcbook.Memory_GIGABYTE,
		},
	}

	laptopClient.SearchLaptop(filter)
}

func testUploadImage(laptopClient *client.LaptopClient) {
	laptop := sample.NewLaptop()
	laptopClient.CreateLaptop(laptop)
	laptopClient.UploadImage(laptop.GetId(), "tmp/laptop.jpg")
}

func authMethods() map[string]bool {
	const laptopServicePath = "/pcbook.LaptopService/"
	return map[string]bool{
		laptopServicePath + "CreateLaptop": true,
		laptopServicePath + "UploadImage":  true,
		laptopServicePath + "RateLaptop":   true,
	}
}

const (
	username        = "admin"
	password        = "admin"
	refreshDuration = 30 * time.Second
)

func loadTLSCreds() (credentials.TransportCredentials, error) {
	// Load CA certificate
	pemServerCA, err := ioutil.ReadFile("certs/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Create creds
	config := &tls.Config{
		RootCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}

func main() {
	serverAdress := flag.String("address", "", "server adress")
	flag.Parse()
	log.Printf("dial server %s", *serverAdress)

	tlsCreds, err := loadTLSCreds()
	if err != nil {
		log.Fatal("cannot load TLS credentials: ", err)
	}

	auth_cc, err := grpc.Dial(*serverAdress, grpc.WithTransportCredentials(tlsCreds))
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}
	auth_client := client.NewAuthClient(auth_cc, username, password)
	interceptor, err := client.NewAuthInterceptor(auth_client, authMethods(), refreshDuration)
	if err != nil {
		log.Fatal("cannot create auth interceptor: ", err)
	}
	laptop_cc, err := grpc.Dial(
		*serverAdress,
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}
	laptopClient := client.NewLaptopClient(laptop_cc)

	// testSearchLaptop(laptopClient)
	// testUploadImage(laptopClient)
	testRateLaptop(laptopClient)
}
