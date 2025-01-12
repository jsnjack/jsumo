package cmd

import "fmt"

const ResetColor = "\033[0m"
const RedColor = "\033[31m"
const GreenColor = "\033[32m"
const YellowColor = "\033[33m"
const BlueColor = "\033[34m"
const PurpleColor = "\033[35m"
const CyanColor = "\033[36m"
const GrayColor = "\033[37m"
const WhiteColor = "\033[97m"
const CrossedColor = "\033[9m"

func red(s interface{}) string {
	return fmt.Sprintf("%s%s%s", RedColor, s, ResetColor)
}

func green(s interface{}) string {
	return fmt.Sprintf("%s%s%s", GreenColor, s, ResetColor)
}

func yellow(s interface{}) string {
	return fmt.Sprintf("%s%s%s", YellowColor, s, ResetColor)
}

func blue(s interface{}) string {
	return fmt.Sprintf("%s%s%s", BlueColor, s, ResetColor)
}

func purple(s interface{}) string {
	return fmt.Sprintf("%s%s%s", PurpleColor, s, ResetColor)
}

func cyan(s interface{}) string {
	return fmt.Sprintf("%s%s%s", CyanColor, s, ResetColor)
}

func gray(s interface{}) string {
	return fmt.Sprintf("%s%s%s", GrayColor, s, ResetColor)
}
