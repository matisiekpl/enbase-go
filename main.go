package main

func main() {
	connectToDatabase()
	connectToPubSub()
	bootstrapRestServer()
	defer changesChannel.Close()
	defer pubsub.Close()
	rest.Logger.Fatal(rest.Start(":1323"))
}
