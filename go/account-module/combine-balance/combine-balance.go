package main

import (
	"fmt"
	"github.com/iotaledger/iota.go/account/builder"
	"github.com/iotaledger/iota.go/account/deposit"
	"github.com/iotaledger/iota.go/account/oracle"
	oracle_time "github.com/iotaledger/iota.go/account/oracle/time"
	"github.com/iotaledger/iota.go/account/store/badger"
	"github.com/iotaledger/iota.go/account/timesrc"
	"github.com/iotaledger/iota.go/api"
	"time"
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
		// Connect to the node
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

	// Get the current time
	now, err := timesource.Time()
	handleErr(err)

	now = now.Add(time.Duration(24) * time.Hour)

	amount, err := account.AvailableBalance()
	handleErr(err)

	// Specify the conditions
	conditions := &deposit.Conditions{TimeoutAt: &now, MultiUse: false, ExpectedAmount: &amount}

	// Generate CDA
	cda, err := account.AllocateDepositAddress(conditions)
	handleErr(err)

	// Set the oracle's threshold to 30 minutes
	threshold := time.Duration(30) * time.Minute
	// Create the oracle
	timeDecider := oracle_time.NewTimeDecider(timesource, threshold)
	// Create a n instance of the oracle
	sendOracle := oracle.New(timeDecider)

	// Ask the oracle if the CDA is OK to send to
	ok, rejectionInfo, err := sendOracle.OkToSend(cda)
	handleErr(err)
	if !ok {
		fmt.Println("Won't send transaction: ", rejectionInfo)
		return
	}

	// Create and send the bundle
	bundle, err := account.Send(cda.AsTransfer())
	handleErr(err)

	fmt.Printf("Sent deposit to %s in the bundle with the following tail transaction hash %s\n", cda.Address, bundle[len(bundle)-1].Hash)

}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
