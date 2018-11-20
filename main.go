package main

func main() {
	connectToDatabase()
	bootstrapRestServer()
	rest.Logger.Fatal(rest.Start(":1323"))
}
