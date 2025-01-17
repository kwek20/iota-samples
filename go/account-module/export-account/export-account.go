package main

import (
	"encoding/json"
	"fmt"
	"os"
	//Store "github.com/iotaledger/iota.go/account/store"
	"github.com/iotaledger/iota.go/account/builder"
	"github.com/iotaledger/iota.go/account/store/badger"
	"github.com/iotaledger/iota.go/account/timesrc"
	"github.com/iotaledger/iota.go/api"
)

// The seed that the account uses to generate CDAs and send bundles
var seed = "PUETTSEITFEVEWCWBTSIZM9NKRGJEIMXTULBACGFRQK9IMGICLBKW9TTEVSDQMGWKBXPVCBMMCXWMNPDX"

func main() {
	// Define the node to connect to
	apiSettings := api.HTTPClientSettings{URI: "https://nodes.devnet.iota.org:443"}
	iotaAPI, err := api.ComposeAPI(apiSettings)
	handleErr(err)

	nodeInfo, err := iotaAPI.GetNodeInfo()
	handleErr(err)
	fmt.Println("Connected to a Devnet node: " + nodeInfo.AppName)

	// Define a database in which to store the seed state
	store, err := badger.NewBadgerStore("seed-state-database")
	handleErr(err)

	// Make sure the database closes when the code stops
	defer store.Close()

	// Use the Google NTP servers as a reliable source of time to check CDA timeouts
	timesource := timesrc.NewNTPTimeSource("time.google.com")

	account, err := builder.NewBuilder().
		// Connect to a node
		WithAPI(iotaAPI).
		// Create the database
		WithStore(store).
		WithSeed(seed).
		// Set the minimum weight magnitude for the Devnet
		WithMWM(9).
		WithTimeSource(timesource).
		// Load the default plugins that enhance the functionality of the account
		WithDefaultPlugins().
		Build()
	handleErr(err)

	handleErr(account.Start())

	// Make sure the account shuts down when the code stops
	defer account.Shutdown()

	// Create a file to which to save the exported seed state
	f, err := os.OpenFile("exported-seed-state.json", os.O_CREATE, 0755)
	handleErr(err)

	// Make sure the file closes when the code stops
	defer f.Close()

	ID := account.ID()

	// Export the seed state
	acc, err := store.ExportAccount(ID)
	handleErr(err)

	// Serialize the seed state as JSON
	jsonacc, err := json.Marshal(acc)
	handleErr(err)

	// Write the seed state to the JSON file
	f.Write(jsonacc)
	f.Close()

	fmt.Println("Seed state exported")

	/* Import code:
	Before you uncomment this code, comment out the export code
	and uncomment the 'Store' package import

	file, err := os.Open("exported-seed-state.json")
	handleErr(err)

	defer file.Close()

	fileinfo, err := file.Stat()
	handleErr(err)

	filesize := fileinfo.Size()
	buffer := make([]byte, filesize)

	jsonSeedState, err := file.Read(buffer)
	handleErr(err)

	a := Store.ExportedAccountState{}
	err = json.Unmarshal(jsonSeedState, &a)
		handleErr(err)

	store.ImportAccount(a)
	*/

}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
