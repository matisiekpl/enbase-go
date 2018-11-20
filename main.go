package main

func main() {
	ConnectToDatabase()
	BootstrapRestServer()
	rest.Logger.Fatal(rest.Start(":1323"))
}
