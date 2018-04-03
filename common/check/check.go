package check

import (
	"fmt"
	"regexp"
)

// 以数字和字母开头,包含下划线和扛
func NumLetterLine(str string) error {
	if str == "" {
		return fmt.Errorf("It is nil")
	}

	if len(str) > 64 {
		return fmt.Errorf("The length of str is invalid, less 64")
	}

	regexpStr := "^[a-zA-Z0-9][a-zA-Z0-9_-]*$"

	regCom, err := regexp.Compile(regexpStr)
	if err != nil {
		return fmt.Errorf("expression of regexp=%v is err: %v", regexpStr, err)
	}

	matchFlag := regCom.MatchString(str)
	if !matchFlag {
		return fmt.Errorf("only start number or char, contain number,char,line")
	}

	return nil
}

// 以数字和字母开头,包含下划线和扛和.
func NumLetterPointLine(str string) error {
	if str == "" {
		return fmt.Errorf("It is nil")
	}

	if len(str) > 64 {
		return fmt.Errorf("The length of str is invalid, less 64")
	}

	regexpStr := "^[a-zA-Z0-9][a-zA-Z0-9._-]*$"

	regCom, err := regexp.Compile(regexpStr)
	if err != nil {
		return fmt.Errorf("expression of regexp=%v is err: %v", regexpStr, err)
	}

	matchFlag := regCom.MatchString(str)
	if !matchFlag {
		return fmt.Errorf("only start number or char, contain number,char,line")
	}

	return nil
}

// 数字和字母
func NumLetter(str string) error {
	if str == "" {
		return fmt.Errorf("It is nil")
	}

	if len(str) > 64 {
		return fmt.Errorf("The length of str is invalid, less 64")
	}

	regexpStr := "^[a-zA-Z0-9]*$"

	regCom, err := regexp.Compile(regexpStr)
	if err != nil {
		fmt.Println("err: ", err)
		return fmt.Errorf("expression of regexp=%v is err: %v", regexpStr, err)
	}

	matchFlag := regCom.MatchString(str)
	if !matchFlag {
		fmt.Println("invalid str: ", str)
		return fmt.Errorf("only start number or char, contain number,char,line")
	}

	return nil
}

// 数字
func NumCheck(str string) error {
	if str == "" {
		return fmt.Errorf("It is nil")
	}

	if len(str) > 16 {
		return fmt.Errorf("The length of num is invalid")
	}

	regexpStr := "^[0-9]*$"

	regCom, err := regexp.Compile(regexpStr)
	if err != nil {
		fmt.Println("err: ", err)
		return fmt.Errorf("expression of regexp=%v is err: %v", regexpStr, err)
	}

	matchFlag := regCom.MatchString(str)
	if !matchFlag {
		fmt.Println("invalid str: ", str)
		return fmt.Errorf("invalid just only num")
	}

	return nil
}
