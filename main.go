// main.go

package main

import (
	"fmt"
	"os"
)

func main() {
	text, err := ConvertPDFToText(os.Args[1], os.Getenv("SNAPSHOT_DIR"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	transactions, err := ExtractTransactionsFromText(text)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	statesJSON, err := ConvertTransactionsToJSON(transactions)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(statesJSON)
}
