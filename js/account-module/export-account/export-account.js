const { createAccount }  = require('@iota/account');
const fs = require('fs');
const ntpClient = require('ntp-client');

const seed = 'PUEOTSEITFEVEWCWBTSIZM9NKRGJEIMXTULBACGFRQK9IMGICLBKW9TTEVSDQMGWKBXPVCBMMCXWMNPDX';

// The node to connect to
const provider = 'https://nodes.devnet.iota.org:443';

// How far back in the Tangle to start the tip selection
const depth = 3;

// The minimum weight magnitude is 9 on the Devnet
const minWeightMagnitude = 9;

// How long to wait before the next attachment round
const delay = 1000 * 30;

// The depth at which transactions are no longer promotable
// Those transactions are automatically re-attached
const maxDepth = 6;

const timeSource = ntpClient.getNetworkTime("time.google.com");

// Create a new account
const account = createAccount({
    seed,
    provider,
    depth,
    minWeightMagnitude,
    delay,
    maxDepth,
    timeSource
});

// Start the plugins
account.start();


account.exportState().then(state => {
    let JSONstate = JSON.stringify(state);
    fs.writeFile('exported-seed-state.json', JSONstate,
    function(err, result) {
        if (err) {
			console.log('error', err);
			// Close the database and stop any ongoing reattachments
            account.stop();
        } else {
            console.log('Seed state saved')
        }
	});
});

/* Import code:
Before you uncomment this code, comment out the export code

let JSONSeedState = fs.readFileSync("exported-seed-state.json");

let importedState = JSON.parse(JSONSeedState);

account.importState(importedState).then(err => {
	if (err) {
		console.log('error', err);
		// Close the database and stop any ongoing reattachments
        account.stop();
	} else {
		console.log('Seed state imported')
	}
});
*/

