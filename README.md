# arango-importer
Golang project to import data from the Firebase-based CDL [client project](https://github.com/canonical-debate-lab/client), hosted at http://canonicaldebate.com, to an ArangoDB database (to be used, for example, by the [server-api](https://github.com/canonical-debate-lab/server-api) project.

## Install ArangoDB
If you haven't already, you need to install the [ArangoDB](https://arangodb.com/) database, version 3.1 or later.

## Install Golang
This project uses the new default dependency management tool that has been available since Golang version 1.11. If you haven't already, go to the [Go Programming Language install page](https://golang.org/doc/install) for instructions on how to get it set up in your environment.

## Install ArangoMiGO
ArangoDB schema creation and migration is managed via the [ArangoMiGO](https://github.com/deusdat/arangomigo) tool.

## Set up this project
If you haven't already, clone this project to a local folder.

## Create the database
First, you must create a local configuration file, as required by ArangoMiGO. An example is located in the file `migrations/config.example` of this project. You can make a copy, and then edit the copy to set your own variables.

```bash
cp migrations/config.example migrations/config
echo "I will use the best editor"
emacs migrations/config
```

At a minimum, you will need to customize the `migrationspath` configuration, and the `username` and `password` in `extras`. You should also change the ArangoDB `username` and `password` for the root user if you didn't use the default configurations.

Next, make sure your ArangoDB server is running:

```bash
/usr/local/opt/arangodb/sbin/arangod
```

Finally, run the ArangoMiGO utility to install the database collections and edges.

```bash
arangomigo migrations/config
```

## Import the data
Once everything is set up, a single command should suffice to convert and import all the data:

```bash
go run *.go
```

This assumes that you just want to import the data that comes with this project, which is a fair-sized graph of two separate debates, one in English on climate change (incomplete), and one in Portuguese regarding a discussion on plans to reform the Brazilian pension fund.

If you'd like to use a different data file, or a different server, you can see the command-line arguments by typing:

```bash
go run *.go --help
```

## Viewing the data
ArangoDB provides two easy ways to interact with the data. They provide out-of-the-box a command line shell:

```bash
arangosh
db._useDatabase("canonical_debate")
db._query("FOR a IN arguments FILTER a.pro == true RETURN a").toArray()
```

Or, you can browse the data using their built-in web interface: http://127.0.0.1:8529/_db/canonical_debate/_admin/aardvark/index.html#graph/debate_map

## Getting the latest data
If you wish to get the most recent dataset from https://canonicaldebate.com, then there are a few options available:

### Download via the Javascript console

Venryx, author of the client project, recommends the following course of actions:

- Open the Canonical Debate site in your browser
- Open the developer / javascript console
- Execute the following commands, one at a time (specifically, the second line will download a LOT of data, and the third line will output it to the console):

```javascript
RR();
let nodes = await GetAsync(()=>RR.GetNodesL2());
JSON.stringify(nodes);
```

- To download that data, press the "Show 2 million more characters" button, then press the Copy button
- Paste the data into a text editor
- Save the data as a file, replacing the current `data/Test1.json` file (or, choose a different name, but then you'll have to change the target filename in the `main.go` file)

### Download the data incrementally using HTTP
You can access a part of this data via any browser using the URL https://firestore.googleapis.com/v1/projects/canonical-debate-prod/databases/(default)/documents/versions/v12-prod/nodes

Unfortunately, you won't be able to retrieve everything: the final element will be a token which allows you to grab the next batch of data:

```json
  "nextPageToken": "AJebfAcovbsCBiFv9bJ2vDZ3SRp8QCHDEJ88P_HNCj1KcXWHJomwWh24E9y1Kwwess-b5CvvIgqNBqYnZE1KKJq2LEyNf36xbAA29VtZtMCy5TTPta5m8DvPKTOHy-nxQKLlSrUseE0CilkDNxj4X4W638-5DpxAkAlW0EDo"
```

This can then be retrieved like so: https://firestore.googleapis.com/v1/projects/canonical-debate-prod/databases/(default)/documents/versions/v12-prod/nodes?pageSize=300&pageToken=NEXT_PAGE_TOKEN_HERE

In the future, this project will offer a command to pull down all the data for you automatically.
