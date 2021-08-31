package serializer_test

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/mikhail-bigun/grpc-app-pcbook/pb/pcbook"
	"github.com/mikhail-bigun/grpc-app-pcbook/sample"
	"github.com/mikhail-bigun/grpc-app-pcbook/serializer"
	"github.com/stretchr/testify/require"
)

func TestFileSerializer(t *testing.T) {
	t.Parallel()

	binaryFile := "../tmp/laptop.bin"
	jsonFile := "../tmp/laptop.json"

	laptop1 := sample.NewLaptop()
	err := serializer.WriteProtobufToBinaryFile(laptop1, binaryFile)
	require.NoError(t, err)

	laptop2 := &pcbook.Laptop{}
	err = serializer.ReadProtobufFromBinaryFile(binaryFile, laptop2)
	require.NoError(t, err)
	require.True(t, proto.Equal(laptop1, laptop2))

	err = serializer.WriteProtobufToJSONFile(laptop1, jsonFile)
	require.NoError(t, err)

}
