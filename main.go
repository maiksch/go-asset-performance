package performance

import "fmt"

// Get asks the yahoo finance API for the performance of the given symbols
func Get(symbols ...[]string) {
	fmt.Println("Getting performancce of symbols", symbols)
}
