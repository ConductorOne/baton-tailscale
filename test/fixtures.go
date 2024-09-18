package test

const FullJSONExample = `// Example/default ACLs for unrestricted connections.
{
	// Declare static groups of users beyond those in the identity service.
	"groups": {
		"group:devs": ["justin.gallardo@insulator.one", "bjorn.tipling@insulator.one", "logan.saso@insulator.one"],
		"group:moredevs": ["michael.burton@insulator.one", "santhosh.kumar@insulator.one", "john.degner@insulator.one"],
	},

	// Declare convenient hostname aliases to use in place of IP addresses.
	"hosts": {
		"example-host-1": "100.100.100.100",
	},

	// Access control lists.
	"acls": [
		// Match absolutely everything.
		// Comment this section out if you want to define specific restrictions.
		{"action": "accept", "users": ["*"], "ports": ["*:*"]},
		{"action": "accept", "src": ["group:security"], "dst": ["john.degner@insulator.one", "logan.saso@insulator.one"]},
		{"action": "accept", "src": ["logan.saso@gmail.com"], "dst": ["*:443", ]},
	],
	"ssh": [
		// Allow all users to SSH into their own devices in check mode.
		// Comment this section out if you want to define specific restrictions.
		{
			"action": "check",
			"src":    ["autogroup:members"],
			"dst":    ["autogroup:self"],
			"users":  ["autogroup:nonroot", "root"],
		},
	],
}
`

const MinimalACLExample = `// Example/default ACLs for unrestricted connections.
{
	"acls": [
		// Allow all users to SSH into their own devices in check mode.
		// Comment this section out if you want to define specific restrictions.
		{
			"action": "accept",
			"src":    ["group:security"],
			"dst":    ["john.degner@insulator.one", "logan.saso@insulator.one"],
		},
	],
}
`

const ExpectedACLResult = `// Example/default ACLs for unrestricted connections.
{
	"acls": [
		// Allow all users to SSH into their own devices in check mode.
		// Comment this section out if you want to define specific restrictions.
		{
			"action": "accept",
			"src":    ["group:security", "bonk.flambe@insulator.one"],
			"dst":    ["john.degner@insulator.one", "logan.saso@insulator.one"],
		},
	],
}
`

const MinimalGroupsExample = `// Example/default ACLs for unrestricted connections.
{
	// Declare static groups of users beyond those in the identity service.
	"groups": {
		// Pre comment
		"group:devs": [
			"justin.gallardo@insulator.one",
			"bjorn.tipling@insulator.one",
			"logan.saso@insulator.one",
		],
		"group:moredevs": [
			"michael.burton@insulator.one",
			"santhosh.kumar@insulator.one",
			"john.degner@insulator.one",
		],
	},
	// Trailing comments
}
`

const ExpectedGroupsResult = `// Example/default ACLs for unrestricted connections.
{
	// Declare static groups of users beyond those in the identity service.
	"groups": {
		// Pre comment
		"group:devs": [
			"justin.gallardo@insulator.one",
			"bjorn.tipling@insulator.one",
			"logan.saso@insulator.one",
			"bonk.flambe@insulator.one",
		],
		"group:moredevs": [
			"michael.burton@insulator.one",
			"santhosh.kumar@insulator.one",
			"john.degner@insulator.one",
		],
	},
	// Trailing comments
}
`

const MinimalSSHExample = `// Example/default ACLs for unrestricted connections.
{
	"ssh": [
		// Allow all users to SSH into their own devices in check mode.
		// Comment this section out if you want to define specific restrictions.
		{
			"action": "check",
			"src":    ["autogroup:members"],
			"dst":    ["autogroup:self"],
			"users":  ["autogroup:nonroot", "root"],
		},
	],
}
`

const ExpectedSSHResult = `// Example/default ACLs for unrestricted connections.
{
	"ssh": [
		// Allow all users to SSH into their own devices in check mode.
		// Comment this section out if you want to define specific restrictions.
		{
			"action": "check",
			"src":    ["autogroup:members", "bonk.flambe@insulator.one"],
			"dst":    ["autogroup:self"],
			"users":  ["autogroup:nonroot", "root"],
		},
	],
}
`
