// preprocessing.go

// This file will contain the preprocessing steps to convert a bank statement pdf into an extracted list of transactions


// 1. Convert the pdf into a text file
// 2. Extract the transactions from the text file
// 3. Convert the transactions into a list of transactions
// 4. Save the list of transactions to a file

// Function Contract:
// Input: pdfFile on computer
// Output: textFile as a string
// Logic: Computer reads the pdf file and converts it to a text file and returns the text file as a string
func ConvertPDFToText(pdfFile string) (string, error) {
	// Step 1: Crop pdf file to only include transactions
	// Inter-Page crop: 2-end
	// Intra-Page crop: crop from the top of the page untill the first red text


	// Step 2: Convert the cropped pdf file to a text file

	// return the text file as a string
	return textFile, nil
}

// Function Contract:
// Input: textFile as a string
// Output: list of transactions
// Logic: Computer reads the text file and extracts the transactions from the text file and returns the list of transactions
func ExtractTransactionsFromText(){
	// Any row that starts with a date is a transaction
	// To extract each field in a transaction, we can delimit by tabs potentially
	// transaction fields are date, description, amount, balance
	// return the list of transactions
	return transactions, nil
}

// Function Contract:
// Input: list of transactions
// Output: JSON file
// Logic: Computer reads the list of transactions and converts it to a JSON file
func ConvertTransactionsToJSON(){
	// For each transaction, create a JSON object where each field is accompanied by a definition of the transaction (TBD)
	// Export the JSON file to the computer
	// return the JSON file as a string
	return JSONFile, nil
}

func main(){

}

