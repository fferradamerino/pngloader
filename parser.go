package pngloader

import (
	"hash/crc32"
)

func parseHeader(file []byte, index int, data PNGData) (int, PNGData) {
	output := data
	expectedHeader := [8]byte{137, 80, 78, 71, 13, 10, 26, 10}

	if len(file) < 8 {
		output.Header = [8]byte{0, 0, 0, 0, 0, 0, 0, 0}
		return 0, output
	}

	for i := 0; i < 8; i++ {
		if file[i + index] != expectedHeader[i] {
			output.IsValid = false
			output.Header = [8]byte{0, 0, 0, 0, 0, 0, 0, 0}
			return 0, output
		}
	}

	output.Header = expectedHeader
	index += 8

	return index, output
}

func parseIHDR(file []byte, index int, data PNGData) (int, PNGData) {
	output := data
	expectedLength := [4]byte{0, 0, 0, 13}
	expectedChunkId := [4]byte{'I', 'H', 'D', 'R'}

	if len(file) < 33 {
		return 0, output
	}

	for i := 0; i < 4; i++ {
		if file[i + index] != expectedLength[i] {
			output.IsValid = false
			return 0, output
		}
	}
	
	for i := 4; i < 8; i++ {
		if file[i + index] != expectedChunkId[i - 4] {
			output.IsValid = false
			return 0, output
		}
	}

	output.Width = uint(file[index + 8]) * 0x1000000 + uint(file[index + 9]) * 0x10000 + uint(file[index + 10]) * 0x100 + uint(file[index + 11])
	output.Height = uint(file[index + 12]) * 0x1000000 + uint(file[index + 13]) * 0x10000 + uint(file[index + 14]) * 0x100 + uint(file[index + 15])

	if output.Width <= 0 || output.Height <= 0 {
		output.IsValid = false
		return 0, output
	}

	output.BitDepth = file[index + 16]

	if output.BitDepth != 1 && output.BitDepth != 2 && output.BitDepth != 4 && output.BitDepth != 8 && output.BitDepth != 16 {
		output.IsValid = false
		return 0, output
	}

	output.ColorType = file[index + 17]

	if output.ColorType != 0 && output.ColorType != 2 && output.ColorType != 3 && output.ColorType != 4 && output.ColorType != 6 {
		output.IsValid = false
		return 0, output
	}

	output.CompressionMethod = file[index + 18]
	output.FilterMethod = file[index + 19]

	if output.CompressionMethod != 0 || output.FilterMethod != 0 {
		output.IsValid = false
		return 0, output
	}

	output.InterlaceMethod = file[index + 20]

	if output.InterlaceMethod != 0 && output.InterlaceMethod != 1 {
		output.IsValid = false
		return 0, output
	}

	expectedChecksum := uint32(file[index + 21]) * 0x1000000 + uint32(file[index + 22]) * 0x10000 + uint32(file[index + 23]) * 0x100 + uint32(file[index + 24])
	if crc32.ChecksumIEEE(file[index + 4 : index + 21]) != expectedChecksum {
		output.IsValid = false
		return 0, output
	}

	return index + 25, output
}

func parseIDAT(file []byte, index int, data PNGData, length uint) (int, PNGData) {
	output := data

	if index + int(length) + 12 > len(file) {
		output.IsValid = false
		return 0, output
	}

	dataChecksum := crc32.ChecksumIEEE(file[index + 4 : index + int(length) + 8])
	expectedChecksum := uint32(file[index + int(length) + 8]) * 0x1000000 + uint32(file[index + int(length) + 9]) * 0x10000 + uint32(file[index + int(length) + 10]) * 0x100 + uint32(file[index + int(length) + 11])
	
	if dataChecksum != expectedChecksum {
		output.IsValid = false
		return 0, output
	}

	for i := index + 8; i < index + int(length) + 8; i++ {
		output.Data = append(output.Data, file[i])
	}

	return index + int(length) + 12, output
}

func parseIEND(file []byte, index int, data PNGData) (int, PNGData) {
	output := data

	if index + 12 != len(file) {
		output.IsValid = false
		return 0, output
	}

	expectedCRC := [4]byte{0xAE, 0x42, 0x60, 0x82}

	for i := 0; i < 4; i++ {
		if file[index + 8 + i] != expectedCRC[i] {
			output.IsValid = false
			return 0, output
		}
	}

	return index + 12, output
}

func parseBKGDtype3(file []byte, index int, data PNGData) PNGData {
	output := data

	if index + 13 <= len(file) {
		output.IsValid = false
		return output
	}

	expectedCRC := uint32(file[index + 9]) * 0x1000000 + uint32(file[index + 10]) * 0x10000 + uint32(file[index + 11]) * 0x100 + uint32(file[index + 12])
	dataChecksum := crc32.ChecksumIEEE(file[index + 9 : index + 13])
	if expectedCRC != dataChecksum {
		output.IsValid = false
		return output
	}
	
	paletteIndex := file[index + 8]
	if int(paletteIndex) >= len(output.Palette) {
		output.IsValid = false
		return output
	}

	output.Background = output.Palette[int(paletteIndex)]
	return output
}

func parseBKGDtype0and4(file []byte, index int, data PNGData) PNGData {
	// TODO
	return data
}

func parseBKGDtype2and6(file []byte, index int, data PNGData) PNGData {
	// TODO
	return data
}

func parseBKGD(file []byte, index int, data PNGData, colorType byte) (int, PNGData) {
	output := data

	switch colorType {
	case 0:
		output = parseBKGDtype0and4(file, index, data)
		return index + 14, output
	case 2:
		output = parseBKGDtype2and6(file, index, data)
		return index + 18, output
	case 3:
		output = parseBKGDtype3(file, index, data)
		return index + 13, output
	case 4:
		output = parseBKGDtype0and4(file, index, data)
		return index + 14, output
	case 6:
		output = parseBKGDtype2and6(file, index, data)
		return index + 18, output
	default:
		output.IsValid = false
		return 0, output
	}
}

func parsePHYS(file []byte, index int, data PNGData) (int, PNGData) {
	output := data

	// TODO

	return index + 21, output
}

func parseTIME(file []byte, index int, data PNGData) (int, PNGData) {
	output := data

	// TODO

	return index + 19, output
}

