package main

import "log"

func main() {
	newStore, err := NewPostgresStorage()

	if err != nil {
		log.Fatalf("There is Something Wrong With Db : %s", err)
	}

	// Initialize The Database
	newStore.Init()

	apiServer := NewAPIServer(":8080", newStore)

	apiServer.Run()
}
