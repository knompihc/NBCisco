/********************************************************************
 * FileName:     lampUtils.go
 * Project:      Havells StreetComm
 * Module:       lampUtils
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/

package lampUtils

func UpdateLampDim(lampDim *int) {
	switch *lampDim {
	case 1, 2:
		*lampDim = 5
	case 3, 4:
		*lampDim = 6
	case 5, 6:
		*lampDim = 7
	case 7, 8:
		*lampDim = 8
	}
}
