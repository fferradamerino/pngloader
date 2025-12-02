package pngloader

import (
	"bytes"
	"compress/zlib"
	"io"
	"os"
)

func loadPNG(file []byte) PNGData {
	var output PNGData
	var index int

	index = 0
	output.IsValid = true
	index, output = parseHeader(file, index, output)

	if !output.IsValid {
		return output
	}
	
	index, output = parseIHDR(file, index, output)

	var currentLength uint
	var currentChunkType string
	for index + 7 < len(file) && output.IsValid {
		currentLength = uint(file[index]) * 0x1000000 + uint(file[index + 1]) * 0x10000 + uint(file[index + 2]) * 0x100 + uint(file[index + 3])
		currentChunkType = string(file[index + 4]) + string(file[index + 5]) + string(file[index + 6]) + string(file[index + 7])

		switch currentChunkType {
		case "IDAT":
			index, output = parseIDAT(file, index, output, currentLength)
		case "IEND":
			index, output = parseIEND(file, index, output)
		/*
		case "PLTE":
			index, output = parsePLTE(file, index, output, currentLength)*/
		case "bKGD":
			index, output = parseBKGD(file, index, output, output.ColorType)
		case "pHYs":
			index, output = parsePHYS(file, index, output)
		case "tIME":
			index, output = parseTIME(file, index, output)
		default:
			output.IsValid = false
			break
		}
	}

	return output
}

func generateRawData(pngdata PNGData) (bool, RawData) {
	var output RawData

	output.Width = uint32(pngdata.Width)
	output.Height = uint32(pngdata.Height)
	output.Format = pngdata.ColorType

	// Stage 1: decompress the data
	b := bytes.NewReader(pngdata.Data)

	r, err := zlib.NewReader(b)
	if err != nil {
		return false, output
	}

	uncompressedData, uncompressError := io.ReadAll(r)

	if uncompressError != nil {
		return false, output
	}

	output.Data = uncompressedData

	r.Close()

	// Stage 2: remove the filtering
	output.Data = removeFiltering(output.Data, output.Width, output.Height, pngdata.BitDepth, output.Format)

	return true, output
}

func LoadPNGAsData(filename string) (bool, RawData) {
	var isCorrect bool
	var rawdata RawData

	file, fileErr := os.ReadFile(filename)

	if fileErr != nil {
		return false, rawdata
	}

	pngdata := loadPNG(file)
	isCorrect, rawdata = generateRawData(pngdata)

	return isCorrect, rawdata
}
