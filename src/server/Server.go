package main

import (
    "fmt"
    "net/http"
    "gopkg.in/mgo.v2"
    "github.com/julienschmidt/httprouter"
    "controllers"
)

/*
* Function main - Setting up httprouter and REST API handlers (GET, POST, PUT) for Trip Planner Service
*/

//var Client Client

func main() {    	
	//Generate OAuth token
	controllers.GetOAuth()
	
    // Instantiate a new router
    r := httprouter.New()
    
    // Controller is going to need a mongo session to use in the CRUD methods. 
    // It would be connected to MongoLab.
    lc := controllers.NewTripController(GetRemoteMGOSession())

	// Add handlers for REST webservices on 'Trip' resource
    // Get all Trip resources
    r.GET("/trips", lc.GetAllTrips)
    
    // Get a Trip resource identified by id
    r.GET("/trips/:id", lc.GetTrip)
	
	// Create a Trip resource
    r.POST("/trips", lc.CreateTrip)
	
	// Update a Trip resource identified by id
	r.PUT("/trips/:id/request", lc.UpdateTrip)
        
    // Fire up the server
    http.ListenAndServe("localhost:7000", r)    
}


//-----------------------Local MongoDb and MongoLab setup----------------------------
/*
* GetMGOSession - used in the server to connect to Local mongodb
*/
func GetMGOSession() *mgo.Session {  
    // Connect to our local mongo
    s, err := mgo.Dial("mongodb://localhost")

    // Check if connection error, is mongo running?
    if err != nil {
    	fmt.Printf("Error : Can't connect to mongo, go error %v\n", err)
        panic(err)
    }
    return s
}


/*
* GetRemoteMGOSession - used in the server to connect to Remote mongolab
*/
func GetRemoteMGOSession() *mgo.Session {
	mongolab_uri := "mongodb://renubatheja:renubatheja@ds035844.mongolab.com:35844/cmpe273-assignment2"
	session, err := mgo.Dial(mongolab_uri)
  	if err != nil {
    	fmt.Printf("Error : Can't connect to mongo, go error %v\n", err)
  	}
	
	session.SetSafe(&mgo.Safe{})
	return session
}

