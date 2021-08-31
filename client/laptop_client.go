package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mikhail-bigun/grpc-app-pcbook/pb/pcbook"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LaptopClient is a client to call laptop service rpcs
type LaptopClient struct {
	service pcbook.LaptopServiceClient
}

// NewLaptopClient returns a new laptop client
func NewLaptopClient(cc *grpc.ClientConn) *LaptopClient {
	service := pcbook.NewLaptopServiceClient(cc)
	return &LaptopClient{
		service: service,
	}
}

// RateLaptop rate a stream of laptops
func (laptopClient *LaptopClient) RateLaptop(laptopIDs []string, scores []float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.service.RateLaptop(ctx)
	if err != nil {
		return fmt.Errorf("cannot rate laptop: %v", err)
	}

	// go routine to receive responses
	waitResponse := make(chan error)
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				log.Print("no more response")
				waitResponse <- nil
				return
			}
			if err != nil {
				waitResponse <- fmt.Errorf("cannot receive stream: %v", err)
			}

			log.Printf("received response: %v", res)
		}
	}()

	// send requests
	for i, laptopID := range laptopIDs {
		req := &pcbook.RateLaptopRequest{
			LaptopId: laptopID,
			Score:    scores[i],
		}

		err := stream.Send(req)
		if err != nil {
			return fmt.Errorf("cannot send stream request: %v - %v", err, stream.RecvMsg(nil))
		}

		log.Print("request sent: ", req)
	}

	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("cannot close stream send: %v", err)
	}

	err = <-waitResponse
	return err
}

// UploadImage upload an image for an existing laptop
func (laptopClient *LaptopClient) UploadImage(laptopID string, imagePath string) {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("cannot open file: ", err)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.service.UploadImage(ctx)
	if err != nil {
		log.Fatal("cannot upload image: ", err)

	}

	req := &pcbook.UploadImageRequest{
		Data: &pcbook.UploadImageRequest_Info{
			Info: &pcbook.ImageInfo{
				LaptopId:  laptopID,
				ImageType: filepath.Ext(imagePath),
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send image info: ", err, stream.RecvMsg(nil))
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk into buffer: ", err)
		}

		req := &pcbook.UploadImageRequest{
			Data: &pcbook.UploadImageRequest_DataChunk{
				DataChunk: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server: ", err, stream.RecvMsg(nil))
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response from server: ", err)
	}

	log.Printf("image uploaded with id: %s, size: %d", res.GetId(), res.GetSize())
}

// SearchLaptop search for a laptops by fiters
func (laptopClient *LaptopClient) SearchLaptop(filter *pcbook.Filter) {
	log.Print("search filter: ", filter)
	// set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pcbook.SearchLaptopRequest{
		Filter: filter,
	}
	stream, err := laptopClient.service.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatal("cannot search laptop: ", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal("cannot recieve response: ", err)
		}

		laptop := res.GetLaptop()
		log.Print("- found: ", laptop.GetId())
		log.Print("		+ brand: ", laptop.GetBrand())
		log.Print("		+ name: ", laptop.GetName())
		log.Print("		+ cpu cores: ", laptop.Cpu.GetNumberOfCores())
		log.Print("		+ cpu min ghz: ", laptop.Cpu.GetMinGhz())
		log.Print("		+ ram: ", laptop.GetRam())
		log.Print("		+ price: ", laptop.GetPriceUsd())
	}
}

// CreateLaptop create a laptop
func (laptopClient *LaptopClient) CreateLaptop(laptop *pcbook.Laptop) {

	req := &pcbook.CreateLaptopRequest{
		Laptop: laptop,
	}

	// set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := laptopClient.service.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Print("laptop already exists")
		} else {
			log.Fatal("cannot create a laptop: ", err)
		}
		return
	}

	log.Printf("created laptop with id: %s", res.Id)
}
