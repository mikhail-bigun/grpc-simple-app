package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/jinzhu/copier"
	"github.com/mikhail-bigun/grpc-app-pcbook/pb/pcbook"
)

// ErrorAlreadyExists is returned when a record with the same ID already present in the store
var ErrorAlreadyExists = errors.New("record already exist")

// LaptopStore is an interface to store laptop
type LaptopStore interface {
	// Save saves the laptop to the store
	Save(laptop *pcbook.Laptop) error

	// Find find a laptop by Id
	Find(id string) (*pcbook.Laptop, error)

	// Search search for a laptops via provided filter, returns one by one via found func
	Search(ctx context.Context, filter *pcbook.Filter, found func(laptop *pcbook.Laptop) error) error
}

// InMemoryLaptopStore stores laptop in memory
type InMemoryLaptopStore struct {
	mutex sync.RWMutex
	data  map[string]*pcbook.Laptop
}

// NewInMemoryLaptopStore returns a new InMemoryLaptopStore
func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pcbook.Laptop),
	}
}

// Save saves the laptop to the store
func (store *InMemoryLaptopStore) Save(laptop *pcbook.Laptop) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.data[laptop.Id] != nil {
		return ErrorAlreadyExists
	}

	// deep copy
	other, err := deepCopy(laptop)
	if err != nil {
		return err
	}

	store.data[other.Id] = other

	return nil
}

// Find finds laptop by provided ID
func (store *InMemoryLaptopStore) Find(id string) (*pcbook.Laptop, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	laptop := store.data[id]
	if laptop == nil {
		return nil, nil
	}

	return deepCopy(laptop)
}

// Search search for a laptops via provided filter, returns one by one via found func
func (store *InMemoryLaptopStore) Search(ctx context.Context, filter *pcbook.Filter, found func(laptop *pcbook.Laptop) error) error {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	for _, laptop := range store.data {

		if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
			log.Print("context is cancelled")
			return errors.New("context is cancelled")
		}

		if isMatchFilter(filter, laptop) {
			// deep copy
			other, err := deepCopy(laptop)
			if err != nil {
				return err
			}

			err = found(other)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func isMatchFilter(filter *pcbook.Filter, laptop *pcbook.Laptop) bool {
	if laptop.GetPriceUsd() > filter.GetMaxPriceUsd() {
		return false
	}
	if laptop.GetCpu().GetNumberOfCores() < filter.GetMinCpuCores() {
		return false
	}
	if laptop.GetCpu().GetMinGhz() < filter.GetMinCpuGhz() {
		return false
	}
	if toBit(laptop.GetRam()) < toBit(filter.GetMinRam()) {
		return false
	}

	return true
}

// toBit convert memory size to the smallest unit
func toBit(memory *pcbook.Memory) uint64 {
	value := memory.GetValue()
	switch memory.GetUnit() {
	case pcbook.Memory_BIT:
		return value
	case pcbook.Memory_BYTE:
		return value << 3 // 8 = 2^3
	case pcbook.Memory_KILOBYTE:
		return value << 13 // 1024*8 = 2^10 * 2^3 = 2^
	case pcbook.Memory_MEGABYTE:
		return value << 23
	case pcbook.Memory_GIGABYTE:
		return value << 33
	case pcbook.Memory_TERABYTE:
		return value << 43
	default:
		return 0
	}
}

func deepCopy(laptop *pcbook.Laptop) (*pcbook.Laptop, error) {
	// deep copy
	other := &pcbook.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data: %w", err)
	}
	return other, nil
}
