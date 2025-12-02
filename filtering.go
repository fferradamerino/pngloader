package pngloader

import (
	"math"
)

func abs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

func paethPredictor(a int, b int, c int) int {
	p := a + b - c
	pa := abs(p - a)
	pb := abs(p - b)
	pc := abs(p - c)

	if pa <= pb && pa <= pc {
		return a
	} else if pb <= pc {
		return b
	} else {
		return c
	}
}

// x is for the pixel in the scanline (1 to width)
// y is for the scanline (1 to height)
// If y = 0, then we are accessing the filtering type of the scanline.
// (1, 3) means the third pixel in the first scanline
// If the coordinates are out of boundaries, then this function will return 0
func getScanlineDataAtXY(data []byte, height uint32, width uint32, x int, y int) byte {
	if !(1 <= uint32(x) && uint32(x) <= width) || !(1 <= uint32(y) && uint32(y) <= height) {
		return 0
	}

	return data[(y - 1) * int(width) + (y - 1) + x]
}

func getRawDataAtXY(rawData []byte, height uint32, width uint32, x int, y int) byte {
	if !(1 <= uint32(x) && uint32(x) <= width) || !(1 <= y && y <= int(height)) {
		return 0
	}

	return rawData[(y - 1) * int(width) + (x - 1)]
}

func getFilteringType(data[] byte, height uint32, width uint32, scanline int) byte {
	if !(1 <= scanline && scanline <= int(height)) {
		return 0
	}

	return data[(scanline - 1) * int(width) + (scanline - 1)]
}

func setRawDataValueAtXY(rawData []byte, height uint32, width uint32, x int, y int, newval byte) []byte {
	if !(1 <= uint32(x) && uint32(x) <= width) || !(1 <= uint32(y) && uint32(y) <= height) {
		return rawData
	}

	rawData[(y - 1) * int(width) + (x - 1)] = newval
	return rawData
}

func noneFilterType(rawData []byte, data []byte, scanline int, width uint32, height uint32) []byte {
	for x := 1; x <= int(width); x++ {
		rawData = setRawDataValueAtXY(rawData, height, width, x, scanline, getScanlineDataAtXY(data, height, width, x, scanline))
	}

	return rawData
}

func subFilterType(rawData []byte, data []byte, scanline int, width uint32, height uint32, bpp int) []byte {
	var dataElement, bppElement byte
	for x := 1; x <= int(width); x++ {
		dataElement = getScanlineDataAtXY(data, height, width, x, scanline)
		bppElement = getRawDataAtXY(rawData, height, width, x - bpp, scanline)
		rawData = setRawDataValueAtXY(rawData, height, width, x, scanline, dataElement + bppElement)
	}
	
	return rawData
}

func upFilterType(rawData []byte, data []byte, scanline int, width uint32, height uint32) []byte {
	var upData, priorData byte
	for x := 1; x <= int(width); x++ {
		upData = getScanlineDataAtXY(data, height, width, x, scanline)
		priorData = getRawDataAtXY(rawData, height, width, x, scanline - 1)
		rawData = setRawDataValueAtXY(rawData, height, width, x, scanline, upData + priorData)
	}

	return rawData
}

func averageFilterType(rawData []byte, data []byte, scanline int, width uint32, height uint32, bpp int) []byte {
	var averageData, rawBppData, priorData, result int
	for x := 1; x <= int(width); x++ {
		averageData = int(getScanlineDataAtXY(data, height, width, x, scanline))
		rawBppData = int(getRawDataAtXY(rawData, height, width, x - bpp, scanline))
		priorData = int(getRawDataAtXY(rawData, height, width, x, scanline - 1))
		result = (averageData + (rawBppData + priorData) / 2) % 256
		rawData = setRawDataValueAtXY(rawData, height, width, x, scanline, byte(result))
	}

	return rawData
}

func paethFilterType(rawData []byte, data []byte, scanline int, width uint32, height uint32, bpp int) []byte {
	var paethData, rawBppData, priorData, priorBppData byte
	var result int
	for x := 1; x <= int(width); x++ {
		paethData = getScanlineDataAtXY(data, height, width, x, scanline)
		rawBppData = getRawDataAtXY(rawData, height, width, x - bpp, scanline)
		priorData = getRawDataAtXY(rawData, height, width, x, scanline - 1)
		priorBppData = getRawDataAtXY(rawData, height, width, x - bpp, scanline - 1)

		result = (int(paethData) + paethPredictor(int(rawBppData), int(priorData), int(priorBppData))) % 256
		
		rawData = setRawDataValueAtXY(rawData, height, width, x, scanline, byte(result))
	}

	return rawData
}

func removeFiltering(data []byte, width uint32, height uint32, bitDepth byte, colorType byte) []byte {
	var pixelSize int
	switch colorType {
	case 0:
		pixelSize = 1
	case 2:
		pixelSize = 3
	case 3:
		pixelSize = 1
	case 4:
		pixelSize = 2
	case 6:
		pixelSize = 4
	}

	rawData := make([]byte, int(width * height) * pixelSize)

	bpp := int(math.Ceil(float64(int(bitDepth) * pixelSize) / 8.0))

	for scanline := 1; scanline <= int(height); scanline++ {
		switch getFilteringType(data, height, width * uint32(pixelSize), scanline) {
		case 0:
			rawData = noneFilterType(rawData, data, scanline, width * uint32(pixelSize), height)
		case 1:
			rawData = subFilterType(rawData, data, scanline, width * uint32(pixelSize), height, bpp)
		case 2:
			rawData = upFilterType(rawData, data, scanline, width * uint32(pixelSize), height)
		case 3:
			rawData = averageFilterType(rawData, data, scanline, width * uint32(pixelSize), height, bpp)
		case 4:
			rawData = paethFilterType(rawData, data, scanline, width * uint32(pixelSize), height, bpp)
		}
	}

	return rawData
}

