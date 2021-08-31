package sample

import (
	"github.com/mikhail-bigun/grpc-app-pcbook/pb/pcbook"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// NewKeyboard generate a new random keyboard
func NewKeyboard() *pcbook.Keyboard {
	keyboard := &pcbook.Keyboard{
		Layout:  randomKeyboardLayout(),
		Backlit: randomBool(),
	}

	return keyboard
}

// NewCPU generate a new random CPU
func NewCPU() *pcbook.CPU {
	brand := randomCPUBrand()
	name := randomCPUName(brand)
	numberOfCores := randomInt(2, 8)
	NumberOfThreads := randomInt(2, 8)
	minGhz := randomFloat64(2.0, 3.5)
	maxGhz := randomFloat64(minGhz, 5.0)
	cpu := &pcbook.CPU{
		Brand:           brand,
		Name:            name,
		NumberOfCores:   uint32(numberOfCores),
		NumberOfThreads: uint32(NumberOfThreads),
		MinGhz:          minGhz,
		MaxGhz:          maxGhz,
	}
	return cpu
}

// NewGPU generate a new random GPU
func NewGPU() *pcbook.GPU {
	brand := randomGPUBrand()
	name := randomGPUName(brand)
	minGhz := randomFloat64(1.0, 1.5)
	maxGhz := randomFloat64(minGhz, 2.0)
	gpu := &pcbook.GPU{
		Brand:  brand,
		Name:   name,
		MinGhz: minGhz,
		MaxGhz: maxGhz,
		Memory: &pcbook.Memory{
			Value: uint64(randomInt(2, 6)),
			Unit:  pcbook.Memory_GIGABYTE,
		},
	}
	return gpu
}

// NewRAM generate a new random RAM
func NewRAM() *pcbook.Memory {
	ram := &pcbook.Memory{
		Value: uint64(randomInt(4, 64)),
		Unit:  pcbook.Memory_GIGABYTE,
	}

	return ram
}

// NewSSD generate a new random SSD
func NewSSD() *pcbook.Storage {
	ssd := &pcbook.Storage{
		Driver: pcbook.Storage_SSD,
		Memory: &pcbook.Memory{
			Value: uint64(randomInt(128, 1024)),
			Unit:  pcbook.Memory_GIGABYTE,
		},
	}

	return ssd
}

// NewHDD generate a new random HDD
func NewHDD() *pcbook.Storage {
	hdd := &pcbook.Storage{
		Driver: pcbook.Storage_HDD,
		Memory: &pcbook.Memory{
			Value: uint64(randomInt(1, 6)),
			Unit:  pcbook.Memory_TERABYTE,
		},
	}

	return hdd
}

// NewScreen generate a new random Screen
func NewScreen() *pcbook.Screen {

	screen := &pcbook.Screen{
		SizeInch:   randomFloat32(13, 17),
		Resolution: randomScreenResolution(),
		Panel:      randomScreenPanel(),
		Multitouch: randomBool(),
	}

	return screen
}

// NewLaptop generate a new random Laptop
func NewLaptop() *pcbook.Laptop {
	brand := randomLaptopBrand()
	laptop := &pcbook.Laptop{
		Id:       randomID(),
		Brand:    randomLaptopBrand(),
		Name:     randomLaptopName(brand),
		Cpu:      NewCPU(),
		Ram:      NewRAM(),
		Gpus:     []*pcbook.GPU{NewGPU()},
		Storages: []*pcbook.Storage{NewHDD(), NewSSD()},
		Screen:   NewScreen(),
		Keyboard: NewKeyboard(),
		Weight: &pcbook.Laptop_WeightKg{
			WeightKg: randomFloat64(1.0, 3.0),
		},
		PriceUsd:    randomFloat64(1500, 3000),
		ReleaseYear: uint32(randomInt(2015, 2021)),
		UpdatedAt:   timestamppb.Now(),
	}
	return laptop
}

// RandomLaptopScore generate a new random laptop score
func RandomLaptopScore() float64 {
	return float64(randomInt(1, 10))
}
