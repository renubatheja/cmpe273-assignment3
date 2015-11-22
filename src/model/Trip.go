package model

import (
	"gopkg.in/mgo.v2/bson"
)

/*
* Trip resource - structure of trip resource
*/
type (
	TripRequest struct {
        StartingFromLocationId 	string       	`json:"starting_from_location_id" bson:"starting_from_location_id"`
        LocationIds				[]string		`json:"location_ids" bson:"location_ids"`
    }
      
    TripResponse struct {
        Id     					bson.ObjectId	`json:"id" bson:"_id"`
        Status   				string        	`json:"status" bson:"status"`
        StartingFromLocationId 	string       	`json:"starting_from_location_id" bson:"starting_from_location_id"`
        BestRouteLocationIds	[]string		`json:"best_route_location_ids" bson:"best_route_location_ids"`
        TotalUberCosts    		float64      	`json:"total_uber_costs" bson:"total_uber_costs"`
        TotalUberDuration    	float64      	`json:"total_uber_duration" bson:"total_uber_duration"`
        TotalDistance 			float64			`json:"total_distance" bson:"total_distance"`
    }    


    TripUpdateResponse struct {
        Id     					bson.ObjectId	`json:"id" bson:"_id"`
        Status   				string        	`json:"status" bson:"status"`
        StartingFromLocationId 	string       	`json:"starting_from_location_id" bson:"starting_from_location_id"`
        NextDestinationLocationId 	string       `json:"next_destination_location_id" bson:"next_destination_location_id"`
        
        BestRouteLocationIds	[]string		`json:"best_route_location_ids" bson:"best_route_location_ids"`
        TotalUberCosts    		float64      	`json:"total_uber_costs" bson:"total_uber_costs"`
        TotalUberDuration    	float64      	`json:"total_uber_duration" bson:"total_uber_duration"`
        TotalDistance 			float64			`json:"total_distance" bson:"total_distance"`		
		UberWaitTimeEta 		int       		 `json:"uber_wait_time_eta" bson:"uber_wait_time_eta"`        
    }    

	//{"driver":null,"eta":8,"location":null,"vehicle":null,"surge_multiplier":1.0}
	RideRequestResponse struct {
        Request_id     			string			`json:"request_id" bson:"request_id"`
        Status   				string        	`json:"status" bson:"status"`
        Driver 					string       	`json:"driver" bson:"driver"`
        Eta 					int       		`json:"eta" bson:"eta"`
        Location				string			`json:"location" bson:"location"`
        Vehicle		    		string      	`json:"vehicle" bson:"vehicle"`
        Surge_multiplier    	float32      	`json:"surge_multiplier" bson:"surge_multiplier"`
    }
    
	//Structure of Error - to be sent if the location service encounters any errors    
    Error struct {
    	Code string `json:"code" bson:"code"`
    	Description string `json:"description" bson:"description"`
    	Fieldname string `json:"fieldname" bson:"fieldname"`
    }
    
    Location struct {
        Id     bson.ObjectId `json:"id" bson:"_id"`
        Name   string        `json:"name" bson:"name"`
        Address string       `json:"address" bson:"address"`
        City    string       `json:"city" bson:"city"`
        State    string      `json:"state" bson:"state"`
        Zipcode string		 `json:"zipcode" bson:"zipcode"`
        Coordinate Coordinate `json:"coordinate" bson:"coordinate"`
    }    
    
	Coordinate struct {   
        	Latitude float64  `json:"lat" bson:"lat"`
        	Longitude float64 `json:"lng" bson:"lng"`
    }
    
)