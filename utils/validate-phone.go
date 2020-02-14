package utils

import (
	"github.com/ttacon/libphonenumber"
)

func VerifyPhone(phone, region string) error {
	if region == "" {
		region = "RU"
	}

	_, err := libphonenumber.Parse(phone, region)
	if err != nil {
		return err
	}

	return nil
}

func ParseE164Phone(number, region string) (string, error) {

	if region == "" {
		region = "RU"
	}

	num, err := libphonenumber.Parse(number, region)
	if err != nil {
		return "", err
	}

	//formattedNum := libphonenumber.Format(num, libphonenumber.NATIONAL)
	strFormattedNum := libphonenumber.Format(num, libphonenumber.E164)
	//numFormatted, err := strconv.Atoi(strFormattedNum)
	//numFormatted, err := strconv.ParseUint(strFormattedNum, 10, 0)


	return strFormattedNum, nil
}
