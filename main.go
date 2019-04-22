package main

var database Database

func main() {
	database = Database{}
	database.Connect()
	hosting := Hosting{
		Database: database,
	}
	go Init()
	hosting.Listen()
}

func Init() {
	var projects []Project
	database.db.Find(&projects)
	for _, project := range projects {
		project.Deploy()
	}
}
