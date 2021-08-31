package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"

	"github.com/google/uuid"
	"github.com/mikhail-bigun/grpc-app-pcbook/pb/pcbook"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// max is 1Mb
const maxImageSize = 1 << 20

// LaptopServer provides laptop services
type LaptopServiceServer struct {
	laptopStore LaptopStore
	imageStore  ImageStore
	ratingStore RatingStore
}

// NewLaptopServer returns a new LaptopServer
func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore, ratingStore RatingStore) *LaptopServiceServer {
	return &LaptopServiceServer{
		laptopStore: laptopStore,
		imageStore:  imageStore,
		ratingStore: ratingStore,
	}
}

// CreateLaptop unary RPC that creates a new Laptop
func (server *LaptopServiceServer) CreateLaptop(ctx context.Context, req *pcbook.CreateLaptopRequest) (*pcbook.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	log.Printf("Recieved a CreateLaptop request with id: %s", laptop.Id)

	if len(laptop.Id) > 0 {
		// check if UUID is Valid
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "laptop ID is not a valid UUID: %v", err)
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot generate a new laptop ID: %v", err)
		}
		laptop.Id = id.String()
	}

	// some heavy processing here
	// time.Sleep(6 * time.Second)

	if err := contextError(ctx); err != nil {
		return nil, err
	}

	// save the laptop to DB
	err := server.laptopStore.Save(laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, ErrorAlreadyExists) {
			code = codes.AlreadyExists
		}
		return nil, status.Errorf(code, "cannot save a laptop to the store: %v", err)
	}

	log.Printf("Saved laptop with id: %s", laptop.Id)

	res := &pcbook.CreateLaptopResponse{
		Id: laptop.Id,
	}

	return res, nil
}

func (server *LaptopServiceServer) SearchLaptop(req *pcbook.SearchLaptopRequest, stream pcbook.LaptopService_SearchLaptopServer) error {
	filter := req.GetFilter()
	log.Printf("recieved a SearchLaptop request with a filter: %v", filter)

	err := server.laptopStore.Search(stream.Context(), filter, func(laptop *pcbook.Laptop) error {
		res := &pcbook.SearchLaptopResponse{
			Laptop: laptop,
		}

		err := stream.Send(res)
		if err != nil {
			return err
		}
		log.Printf("sent laptop with id: %s", laptop.GetId())
		return nil
	})

	if err != nil {
		return status.Errorf(codes.Internal, "unexpected error: %v", err)
	}

	return nil
}

// UploadImage upload image by a stream of byte data
func (server *LaptopServiceServer) UploadImage(stream pcbook.LaptopService_UploadImageServer) error {

	req, err := stream.Recv()
	if err != nil {
		log.Print("cannot recieve image info", err)
		return status.Errorf(codes.Unknown, "cannot recieve image info")
	}

	laptopID := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()

	log.Printf("recieved an UploadImage request for laptop %s with image type %s", laptopID, imageType)

	laptop, err := server.laptopStore.Find(laptopID)
	if err != nil {
		log.Print("cannot find laptop", err)
		return status.Errorf(codes.Internal, "cannot find laptop: %v", err)
	}
	if laptop == nil {
		log.Print("laptop does not exists", err)
		return status.Errorf(codes.Internal, "laptop does not exists: %v", err)
	}

	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		if err := contextError(stream.Context()); err != nil {
			return err
		}
		log.Print("waiting for data to be received")

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			log.Print("cannot receive data chunk", err)
			return status.Errorf(codes.Unknown, "cannot receive data chunk: %v", err)
		}

		chunk := req.GetDataChunk()
		size := len(chunk)

		log.Printf("received chunk of data with size: %d", size)

		imageSize += size
		if imageSize > maxImageSize {
			log.Printf("image size is to large: %d > %d", imageSize, maxImageSize)
			return status.Errorf(codes.InvalidArgument, "image size is to large: %d > %d", imageSize, maxImageSize)
		}

		_, err = imageData.Write(chunk)
		if err != nil {
			log.Print("cannot write chunk of data", err)
			return status.Errorf(codes.Internal, "cannot write chunk of data: %v", err)
		}
	}

	imageID, err := server.imageStore.Save(laptopID, imageType, imageData)
	if err != nil {
		log.Print("cannot save image to the store", err)
		return status.Errorf(codes.Internal, "cannot save image to the store: %v", err)
	}

	res := &pcbook.UploadImageResponse{
		Id:   imageID,
		Size: uint32(imageSize),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		log.Print("cannot send response", err)
		return status.Errorf(codes.Unknown, "cannot send response: %v", err)
	}

	log.Printf("image saved with id: %s, size: %d", imageID, imageSize)

	return nil
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		log.Print("request is canceled")
		return status.Error(codes.Canceled, "request is canceled")
	case context.DeadlineExceeded:
		log.Print("deadline is exceeded")
		return status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	default:
		return nil
	}
}

// RateLaptop bidirectional stream that allows client to rate a stream of laptops with a score and returns a stream of average score for each of them
func (server *LaptopServiceServer) RateLaptop(stream pcbook.LaptopService_RateLaptopServer) error {
	
	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Printf("no more data")
			break
		}
		if err != nil {
			log.Printf("cannot receive stream request: %v", err)
			return status.Errorf(codes.Internal, "cannot receive stream request: %v", err)
		}

		laptopID := req.GetLaptopId()
		score := req.GetScore()

		log.Printf("received a rate-laptop request: id = %s, score = %.2f", laptopID, score)

		found, err := server.laptopStore.Find(laptopID)
		if err != nil {
			log.Printf("cannot find laptop in store: %v", err)
			return status.Errorf(codes.Internal, "cannot find laptop in store: %v", err)
		}
		if found == nil {
			return status.Errorf(codes.NotFound, "laptopID: %s not found", laptopID) 
		}

		rating, err := server.ratingStore.Add(laptopID, score)
		if err != nil {
			log.Printf("cannot add rating to the store: %v", err)
			return status.Errorf(codes.Internal, "cannot add rating to the store: %v", err)
		}

		res := &pcbook.RateLaptopResponse{
			LaptopId:     laptopID,
			TimesRated: rating.Count,
			AverageScore: rating.Sum / float64(rating.Count),
		}

		err = stream.Send(res)
		if err != nil {
			log.Printf("cannot send stream response: %v", err)
			return status.Errorf(codes.Unknown, "cannot send stream response: %v", err)
		}
	}

	return nil
}
